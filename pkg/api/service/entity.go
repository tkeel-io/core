package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tkeel-io/core/pkg/service"
	"github.com/tkeel-io/core/utils"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
)

const (
	//
	//	EmptyEntityId = ""
	resultNull = "null"

	//entity table fields.
	entityFieldId             = "id"
	entityFieldUserId         = "user_id"
	entityFieldSource         = "source"
	entityFieldTag            = "tag"
	entityFieldStatus         = "status"
	entityFieldVersion        = "version"
	entityFieldDeletedId      = "deleted_id"
	entityStateFieldPrefix    = "__internal_"
	entityDeleteIdFieldPrefix = "deleted_"

	//entity status
	//	entityStatusActive   = "active"
	entityStatusDeactive = "deactive"
	entityStatusDeleted  = "deleted"

	kvPair    = "%s='%s'"
	whereText = "where id='%s' and user_id='%s' and source='%s' and status != 'deleted'"
	//	entityGetSql    = "select * from %s %s"
	entityUpdateSql = "update %s set %s, version=version+1 %s"
	entityDeleteSql = "update %s set %s %s"
	entityExistsSql = "select 1 from %s %s"
	entityCreateSql = "insert into %s (id, user_id, source, tag, status, version, entity_key) values('%s', '%s', '%s', '%s', '%s', %d, '%s')"
)

var (
	//	errEntityIdRequired = entityFieldRequired("entityId")
	errBodyMustBeJson = errors.New("body must be json(kv).")
	errEntityNotExist = errors.New("entity not exists.")
	errEntityInternal = errors.New("entity internal error.")
)

func entityExisted(entityId string) error {
	return fmt.Errorf("entity(%s)  exised.", entityId)
}

func entityFieldRequired(fieldName string) error {
	return fmt.Errorf("entity field(%s) required.", fieldName)
}

func internalFieldName(fieldName string) string {
	return fmt.Sprintf("%s%s", entityStateFieldPrefix, fieldName)
}

type Entity struct {
	Id      string                 `json:"id"`
	Tag     string                 `json:"tag"`
	Type    string                 `json:"type"`
	Source  string                 `json:"source"`
	UserId  string                 `json:"user_id"`
	Version int64                  `json:"version"`
	KValues map[string]interface{} `json:"kvalues"`
}

type EntityServiceConfig struct {
	TableName   string
	StateName   string
	BindingName string
}

// EntityService is a time-series service.
type EntityService struct {
	daprClient  dapr.Client
	tableName   string
	stateName   string
	bindingName string
}

// NewEntityService returns a new EntityService
func NewEntityService(entityConfig *EntityServiceConfig) (*EntityService, error) {

	cli, err := dapr.NewClient()

	return &EntityService{
		daprClient:  cli,
		tableName:   entityConfig.TableName,
		stateName:   entityConfig.StateName,
		bindingName: entityConfig.BindingName,
	}, err
}

// Name return the name.
func (this *EntityService) Name() string {
	return "entity"
}

// RegisterService register some methods
func (this *EntityService) RegisterService(daprService common.Service) error {
	//register all handlers.
	if err := daprService.AddServiceInvocationHandler("entities", this.entityHandler); nil != err {
		return err
	} else if err = daprService.AddServiceInvocationHandler("entitylist", this.entityList); nil != err {
		return err
	}
	return nil
}

// Echo test for RegisterService.
func (this *EntityService) entityHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}

	log.Info("call entity handler.", in.Verb, in.QueryString, in.DataTypeURL, string(in.Data))

	switch in.Verb {
	case http.MethodGet:
		return this.entityGet(ctx, in)
	case http.MethodPost:
		return this.entityUpsert(ctx, in)
	case http.MethodPut:
		return this.entityUpdate(ctx, in)
	case http.MethodDelete:
		return this.entityDelete(ctx, in)
		// 临时
	case http.MethodPatch:
		return this.entityCreate(ctx, in)
	default:
	}

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func (this *EntityService) getValFromValues(values url.Values, key string) (string, error) {

	if vals, exists := values[key]; exists && len(vals) > 0 {
		return vals[0], nil
	}

	return "", entityFieldRequired(key)
}

func getStringFrom(ctx context.Context, key string) (string, error) {
	if val := ctx.Value(key); nil != val {
		return val.(string), nil
	}
	return "", entityFieldRequired(key)
}

func (this *EntityService) getEntityFrom(ctx context.Context, in *common.InvocationEvent, entityIdRequired bool) (entity Entity, err error) {

	var values url.Values

	if values, err = url.ParseQuery(in.QueryString); nil != err {
		return
	}

	//source field
	if entity.Source, err = getStringFrom(ctx, service.HeaderSource); nil == err {
		log.Info("parse http request field(source) from header successed.")
	} else if entity.Source, err = this.getValFromValues(values, entityFieldSource); nil != err {
		log.Error("parse http request field(source) from query failed", ctx, err)
		return
	}

	//userId field
	if entity.UserId, err = getStringFrom(ctx, service.HeaderUser); nil == err {
		log.Info("parse http request field(user) from header successed.")
	} else if entity.UserId, err = this.getValFromValues(values, entityFieldUserId); nil != err {
		log.Error("parse http request field(user) from query failed", ctx, err)
		return
	}

	//entity id field
	if entity.Id, err = this.getValFromValues(values, entityFieldId); nil != err {
		if !entityIdRequired {
			err = nil
			entity.Id = utils.GenerateUUID()
		}
	}

	//tags
	if vals, exists := values[entityFieldTag]; exists && len(vals) > 0 {
		entity.Tag = strings.Join(vals, ";")
	}

	return entity, checkRequest(entity)
}

func checkRequest(entity Entity) error {
	if "" == entity.Source {
		return entityFieldRequired(entityFieldSource)
	}
	return nil
}

func (this *EntityService) entityExists(ctx context.Context, source, userId, entityId string) error {

	var (
		err     error
		result  *dapr.BindingEvent
		sqlText = fmt.Sprintf(entityExistsSql, this.tableName,
			fmt.Sprintf(whereText, entityId, userId, source))
	)

	if result, err = this.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
		Name:      this.bindingName,
		Operation: "query",
		Metadata: map[string]string{
			"sql": sqlText,
		},
	}); nil != err {
		return err
	} else if resultNull == string(result.Data) {
		err = errEntityNotExist
	}

	return err
}

//EntityGet returns a entity information.
func (this *EntityService) entityGet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	var (
		entity    Entity
		stateItem *dapr.StateItem
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = this.getEntityFrom(ctx, in, true); nil != err {
		return
	} else if err = this.entityExists(ctx, entity.Source, entity.UserId, entity.Id); nil != err {
		log.Error("call entity.Exists failed. ", err)
		return
	} else if stateItem, err = this.daprClient.GetState(ctx, this.stateName, entity.Id); nil == err {
		out.Data = stateItem.Value
	}

	return
}

//EntityGet create  an entity.
func (this *EntityService) entityCreate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	var (
		entity  Entity
		kvalues = make(map[string]interface{})
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = this.getEntityFrom(ctx, in, false); nil != err {
		return
	} else if err = this.entityExists(ctx, entity.Source, entity.UserId, entity.Id); nil == err {
		err = entityExisted(entity.Id)
		return
	}

	sqlText := fmt.Sprintf(entityCreateSql, this.tableName,
		entity.Id, entity.UserId, entity.Source, entity.Tag, entityStatusDeactive, entity.Version, entity.Id)

	//insert entity to binding
	if _, err = this.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
		Name:      this.bindingName,
		Operation: "exec",
		Metadata: map[string]string{
			"sql": sqlText,
		},
	}); nil != err {
		return
	}

	if len(in.Data) > 0 {
		if err = json.Unmarshal(in.Data, &kvalues); nil != err {
			return out, errBodyMustBeJson
		}
	}

	kvalues[internalFieldName(entityFieldTag)] = entity.Tag
	kvalues[internalFieldName(entityFieldId)] = entity.Id
	kvalues[internalFieldName(entityFieldUserId)] = entity.UserId
	kvalues[internalFieldName(entityFieldSource)] = entity.Source
	kvalues[internalFieldName(entityFieldVersion)] = entity.Version

	//encode kvs.
	if out.Data, err = json.Marshal(kvalues); nil != err {
		return
	}

	//save entity state.
	if err = this.daprClient.SaveState(ctx, this.stateName, entity.Id, out.Data); nil != err {
		//redo binding...
		return
	}

	return
}

//EntityGet update an entity.
func (this *EntityService) entityUpdate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	var (
		entity    Entity
		stateItem *dapr.StateItem
		kvalues   = make(map[string]interface{})
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = this.getEntityFrom(ctx, in, true); nil != err {
		return
	} else if err = this.entityExists(ctx, entity.Source, entity.UserId, entity.Id); nil != err {
		return
	}

	if "" != entity.Tag {

		sqlText := fmt.Sprintf(entityUpdateSql, this.tableName, fmt.Sprintf(kvPair, entityFieldTag, entity.Tag),
			fmt.Sprintf(whereText, entity.Id, entity.UserId, entity.Source))

		//update entity to binding
		if _, err = this.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
			Name:      this.bindingName,
			Operation: "exec",
			Metadata: map[string]string{
				"sql": sqlText,
			},
		}); nil != err {
			return
		}
	}

	//get entity from state.
	if stateItem, err = this.daprClient.GetState(ctx, this.stateName, entity.Id); nil == err {
		if err = json.Unmarshal(stateItem.Value, &kvalues); nil != err {
			return out, errEntityInternal
		}
	}

	if len(in.Data) > 0 {
		if err = json.Unmarshal(in.Data, &kvalues); nil != err {
			return out, errBodyMustBeJson
		}
	}

	if "" != entity.Tag {
		kvalues[internalFieldName(entityFieldTag)] = entity.Tag
	}

	if out.Data, err = json.Marshal(kvalues); nil != err {
		return
	} else if err = this.daprClient.SaveState(ctx, this.stateName, entity.Id, out.Data); nil != err {
		//redo binding...
		fmt.Println("TODO")
	}

	return
}

//entityUpsert
func (this *EntityService) entityUpsert(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	var (
		entity    Entity
		sqlText   string
		stateItem *dapr.StateItem
		kvalues   = make(map[string]interface{})
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = this.getEntityFrom(ctx, in, false); nil != err {
		return
	} else if err = this.entityExists(ctx, entity.Source, entity.UserId, entity.Id); nil != err {
		//create entity if not exists.
		if errEntityNotExist != err {
			return
		}
		sqlText = fmt.Sprintf(entityCreateSql, this.tableName,
			entity.Id, entity.UserId, entity.Source, entity.Tag, entityStatusDeactive, entity.Version, entity.Id)

		kvalues[internalFieldName(entityFieldTag)] = entity.Tag
		kvalues[internalFieldName(entityFieldId)] = entity.Id
		kvalues[internalFieldName(entityFieldUserId)] = entity.UserId
		kvalues[internalFieldName(entityFieldSource)] = entity.Source
		kvalues[internalFieldName(entityFieldVersion)] = entity.Version
	} else {
		//update entity if aready exists.
		if "" != entity.Tag {
			sqlText = fmt.Sprintf(entityUpdateSql, this.tableName, fmt.Sprintf(kvPair, entityFieldTag, entity.Tag),
				fmt.Sprintf(whereText, entity.Id, entity.UserId, entity.Source))
		}

		//get entity from state.
		if stateItem, err = this.daprClient.GetState(ctx, this.stateName, entity.Id); nil == err {
			if err = json.Unmarshal(stateItem.Value, &kvalues); nil != err {
				return out, errEntityInternal
			}
		}
	}

	if len(sqlText) > 0 {
		//upsert entity to binding
		if _, err = this.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
			Name:      this.bindingName,
			Operation: "exec",
			Metadata: map[string]string{
				"sql": sqlText,
			},
		}); nil != err {
			return
		}
	}

	if len(in.Data) > 0 {
		if err = json.Unmarshal(in.Data, &kvalues); nil != err {
			return out, errBodyMustBeJson
		}
	}

	if "" != entity.Tag {
		kvalues[internalFieldName(entityFieldTag)] = entity.Tag
	}

	if out.Data, err = json.Marshal(kvalues); nil != err {
		return
	} else if err = this.daprClient.SaveState(ctx, this.stateName, entity.Id, out.Data); nil != err {
		//redo binding...
		fmt.Println("TODO")
	}

	return
}

func generateDeletedId(entityId string) string {
	id := entityDeleteIdFieldPrefix + entityId + utils.GenerateUUID()
	if len(id) > 127 {
		id = id[:127]
	}
	return id
}

//EntityGet delete an entity.
func (this *EntityService) entityDelete(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	var entity Entity

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = this.getEntityFrom(ctx, in, true); nil != err {
		return
	}

	setText := strings.Join([]string{
		fmt.Sprintf(kvPair, entityFieldDeletedId, entity.Id),
		fmt.Sprintf(kvPair, entityFieldStatus, entityStatusDeleted),
		fmt.Sprintf(kvPair, entityFieldId, generateDeletedId(entity.Id)),
	}, ",")

	sqlText := fmt.Sprintf(entityDeleteSql, this.tableName, setText,
		fmt.Sprintf(whereText, entity.Id, entity.UserId, entity.Source))

	//delete entity to binding
	if _, err = this.daprClient.InvokeBinding(ctx, &dapr.InvokeBindingRequest{
		Name:      this.bindingName,
		Operation: "exec",
		Metadata: map[string]string{
			"sql": sqlText,
		},
	}); nil != err {
		return
	}

	fmt.Println("delete entity", sqlText, err)

	//delete entity state.
	if err = this.daprClient.DeleteState(ctx, this.stateName, entity.Id); nil != err {
		//redo binding...
		return
	}

	return
}

// Echo test for RegisterService.
func (this *EntityService) entityList(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {

	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}

	//parse request query...

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}
