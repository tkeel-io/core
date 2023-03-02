package runtime

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/scheme"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/core/third_party/jsonpatch"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"github.com/tkeel-io/tdtl/pkg/json/jsonparser"
)

// some persistent field enumerate.
const (
	FieldID          string = "id"
	FieldType        string = "type"
	FieldOwner       string = "owner"
	FieldSource      string = "source"
	FieldVersion     string = "version"
	FieldLastTime    string = "last_time"
	FieldTemplate    string = "template_id"
	FieldScheme      string = "scheme"
	FieldDescription string = "description"
	FieldProperties  string = "properties"
	FieldRawData     string = "properties.rawData"
	FieldKeyWords    string = "search_model"
	// FieldEntitySource string = "entity_source".

)

var schemeCache = NewNodeCache()

type PathConstructor func(pc v1.PathConstructor, destVal, setVal []byte, path string) ([]byte, string, error)

type entity struct {
	id              string
	state           tdtl.Collect
	pathConstructor PathConstructor
}

func DefaultEntity(id string) Entity {
	return &entity{id: id, state: *tdtl.New([]byte(`{"properties":{}}`)), pathConstructor: pathConstructor}
}

func NewEntity(id string, state []byte) (Entity, error) {
	s := tdtl.New(state)
	// construct scheme if not exists.
	scheme := s.Get(FieldScheme)
	if tdtl.Null == scheme.Type() {
		s.Set(FieldScheme, tdtl.New([]byte("{}")))
	}

	return &entity{
			id:              id,
			state:           *s,
			pathConstructor: pathConstructor,
		},
		errors.Wrap(s.Error(), "new entity")
}

func (e *entity) ID() string {
	return e.id
}

func (e *entity) Get(path string) tdtl.Node {
	if isFieldScheme(path) {
		ret, ok := schemeCache.Get(e.id, path)
		if ok {
			return ret
		}

		ret = e.state.Get(path)
		schemeCache.Set(e.id, path, ret)
		return ret
	}
	return e.state.Get(path)
}

func (e *entity) Handle(ctx context.Context, feed *Feed) *Feed { //nolint
	if nil != feed.Err {
		return feed
	}

	changes := []Patch{}
	pc := feed.Event.Attr(v1.MetaPathConstructor)

	cc := e.state.Copy()
	cleanSchemaCache := false
	for _, patch := range feed.Patches {
		if isFieldScheme(patch.Path) {
			cleanSchemaCache = true
		}
		switch patch.Op {
		case xjson.OpAdd:
			cc.Append(patch.Path, patch.Value)
		case xjson.OpCopy:
		case xjson.OpMerge:
			var err error
			if patch.Value.Type() != tdtl.Null {
				err = merge(cc, patch, e, feed)
				if err != nil {
					return feed
				}
			} else {
				log.L().Error("merge entity error, patch value is null", logf.Eid(e.id), logf.Error(cc.Error()),
					logf.Any("patches", feed.Patches), logf.Event(feed.Event))
			}
		case xjson.OpRemove:
			cc.Del(patch.Path)
		case xjson.OpReplace:
			// construct sub path if not exists.
			pcIns := v1.PathConstructor(pc)
			if patch.Value.Type() != tdtl.Null {
				patchVal, patchPath, err := e.pathConstructor(pcIns, cc.Raw(), patch.Value.Raw(), patch.Path)
				if nil != err {
					log.L().Error("update entity", logf.Eid(e.id), logf.Error(err),
						logf.Any("patches", feed.Patches), logf.Event(feed.Event))
					// in.Patches 处理完毕，丢弃.
					feed.Err = err
					feed.Patches = []Patch{}
					feed.State = e.Raw()
					return feed
				}
				cc.Set(patchPath, tdtl.New(patchVal))
			} else {
				log.L().Error("replace entity error, patch value is null", logf.Eid(e.id), logf.Error(cc.Error()),
					logf.Any("patches", feed.Patches), logf.Event(feed.Event))
			}
		default:
			return &Feed{Err: xerrors.ErrPatchPathInvalid}
		}

		if nil != cc.Error() {
			log.L().Error("update entity", logf.Eid(e.id), logf.Error(cc.Error()),
				logf.Any("patches", feed.Patches), logf.Event(feed.Event))
			break
		}

		switch patch.Op {
		case xjson.OpMerge:
			patch.Value.Foreach(func(key []byte, value *tdtl.Collect) {
				changes = append(changes, Patch{
					Op: xjson.OpReplace, Value: value,
					Path: strings.Join([]string{patch.Path, string(key)}, "."),
				})
			})
		default:
			changes = append(changes,
				Patch{Op: patch.Op, Path: patch.Path, Value: patch.Value})
		}
	}

	if cc.Error() == nil {
		e.state = *cc
		version := e.Version()
		if version%100 == 99 {
			e.cleanTelemetry()
		}
		if cleanSchemaCache {
			schemeCache.Delete(e.id)
		}
		e.Update()
	} else {
		log.L().Error("update entity", logf.Error(cc.Error()), logf.Eid(e.id),
			logf.Event(feed.Event), logf.Value(feed.Patches))
	}

	// in.Patches 处理完毕，丢弃.
	feed.Err = cc.Error()
	feed.Changes = changes
	feed.Patches = []Patch{}
	feed.State = e.Raw()
	return feed
}

func merge(cc *tdtl.JSONNode, patch Patch, e Entity, feed *Feed) error {
	tc := cc.Get(patch.Path)
	if tc.Type() == tdtl.Null {
		cc.Set(patch.Path, patch.Value)
		return nil
	}
	if tc.Type() != tdtl.Object && tc.Type() != tdtl.Object {
		feed.Err = errors.New("datatype is not object")
		feed.Patches = []Patch{}
		feed.State = e.Raw()
		return feed.Err
	}
	ntc, err := jsonpatch.MergePatch(tc.Raw(), patch.Value.Raw())
	if err != nil {
		feed.Err = errors.New("datatype is not object")
		feed.Patches = []Patch{}
		feed.State = e.Raw()
		return feed.Err
	}

	cc.Set(patch.Path, tdtl.New(ntc))
	return nil
}

func merge1(cc *tdtl.JSONNode, patch Patch, e Entity, feed *Feed) error {
	mval := cc.Get(patch.Path).Merge(patch.Value)
	err := mval.Error()
	if tdtl.Object != mval.Type() {
		log.Error("patch merge", logf.Eid(e.ID()), logf.Error(err),
			logf.Any("patches", feed.Patches), logf.Event(feed.Event))
		err = xerrors.ErrInternal
	}

	if nil != err {
		feed.Err = err
		feed.Patches = []Patch{}
		feed.State = e.Raw()
		return err
	}
	cc.Set(patch.Path, mval)
	return nil
}

func (e *entity) Raw() []byte {
	return e.state.Copy().Raw()
}

func (e *entity) Copy() Entity {
	cp := e.state.Copy()
	return &entity{
		id:    e.id,
		state: *cp,
	}
}

func (e *entity) Basic() *tdtl.Collect {
	basic := e.state.Copy()
	basic.Set("scheme", tdtl.New([]byte("{}")))
	basic.Set("properties", tdtl.New([]byte("{}")))
	return basic
}

func (e *entity) Tiled() tdtl.Node {
	basic := e.state.Copy()
	basic.Del(FieldScheme)
	basic.Del(FieldProperties)
	result := basic.Merge(tdtl.New(e.Properties().Raw()))
	return result
}

func (e *entity) Type() string {
	return e.state.Get(FieldType).String()
}

func (e *entity) Owner() string {
	return e.state.Get(FieldOwner).String()
}

func (e *entity) Source() string {
	return e.state.Get(FieldSource).String()
}

func (e *entity) Version() int64 {
	version := e.state.Get(FieldVersion).String()
	i, _ := strconv.ParseInt(version, 10, 64)
	return i
}

func (e *entity) LastTime() int64 {
	lastTime := e.state.Get(FieldLastTime).String()
	i, _ := strconv.ParseInt(lastTime, 10, 64)
	return i
}

func (e *entity) TemplateID() string {
	return e.state.Get(FieldTemplate).String()
}

func (e *entity) Properties() tdtl.Node {
	return e.state.Get("properties")
}

func (e *entity) Scheme() tdtl.Node {
	return e.state.Get("scheme")
}

func (e *entity) GetProp(key string) tdtl.Node {
	return e.state.Get("properties." + key)
}

func (e *entity) Update() {
	// update entity version.
	lastTime := time.Now().UnixNano() / 1e6
	e.state.Set(FieldVersion, tdtl.NewInt64(e.Version()+1))
	// update entity last_time.
	e.state.Set(FieldLastTime, tdtl.NewInt64(lastTime))
}

func (e *entity) cleanTelemetry() {
	delKeys := map[string]bool{}
	tdtl.New(e.GetProp("telemetry").Raw()).
		Foreach(func(key []byte, value *tdtl.Collect) {
			delKeys[string(key)] = true
		})
	tdtl.New(e.Get("scheme.telemetry.define.fields").Raw()).
		Foreach(func(key []byte, value *tdtl.Collect) {
			delKeys[string(key)] = false // don't delete
		})

	cc := e.state.Copy()
	for k, del := range delKeys {
		if del {
			cc.Del("properties.telemetry." + k)
		}
	}
	e.state = *cc
}

func pathConstructor(pc v1.PathConstructor, destVal, setVal []byte, path string) (_ []byte, _ string, err error) {
	switch pc {
	case v1.PCScheme:
		setVal, path, err = makeSubPath(destVal, setVal, path)
		return setVal, path, errors.Wrap(err, "make sub path")
	default:
	}
	return setVal, path, nil
}

func makeSubPath(dest, src []byte, path string) ([]byte, string, error) {
	var index int
	segs := strings.Split(path, ".")
	seg0, segs := segs[0], segs[1:]
	dest = tdtl.New(dest).Get(seg0).Raw()
	for ; index < len(segs); index += 3 {
		if _, _, _, err := jsonparser.Get(dest, segs[:index+1]...); nil != err {
			if errors.Is(err, jsonparser.KeyPathNotFoundError) {
				break
			}
			return nil, path, errors.Wrap(err, "make sub path")
		}
	}

	if index >= len(segs) {
		return src, path, nil
	}

	if missSegs := segs[index:]; len(missSegs) > 3 {
		path = strings.Join(append([]string{seg0}, segs[:index+1]...), ".")
		return makeScheme(missSegs, src), path, nil
	}
	return src, path, nil
}

func makeScheme(segs []string, data []byte) []byte {
	cfg := &scheme.Config{
		ID:                segs[0],
		Type:              "struct",
		Name:              segs[0],
		Enabled:           true,
		EnabledSearch:     true,
		EnabledTimeSeries: true,
		Define:            map[string]interface{}{},
		LastTime:          time.Now().UnixNano() / 1e6,
	}

	head := cfg
	mids := segs[3 : len(segs)-1]
	for index := 0; index < len(mids); index += 3 {
		curCfg := &scheme.Config{
			ID:                mids[index],
			Type:              "struct",
			Name:              mids[index],
			Enabled:           true,
			EnabledSearch:     true,
			EnabledTimeSeries: true,
			Define:            map[string]interface{}{},
			LastTime:          time.Now().UnixNano() / 1e6,
		}

		head.Define["fields"] = map[string]interface{}{mids[index]: curCfg}
		head = curCfg
	}

	// set last seg.
	var v interface{}
	json.Unmarshal(data, &v)
	head.Define["fields"] = map[string]interface{}{
		segs[len(segs)-1]: v,
	}
	bytes, _ := json.Marshal(cfg)

	return bytes
}

func isFieldScheme(path string) bool {
	return strings.Index(path, FieldScheme) == 0
}
