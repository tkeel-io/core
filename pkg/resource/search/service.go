package search

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/resource/search/driver"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

var GlobalService *Service

func Init() {
	defaultRegistered := map[driver.Type]driver.SearchEngine{
		// Add other drivers to SearchService here.
		driver.Elasticsearch: driver.NewElasticsearchEngine(config.Get().SearchEngine.ES),
	}
	GlobalService = NewService(defaultRegistered).SetSelectOptions(driver.WithElasticsearch)
}

var _ pb.SearchHTTPServer = &Service{}

type Service struct {
	drivers    map[driver.Type]driver.SearchEngine
	selectOpts []driver.SelectDriveOption
}

func NewService(registered map[driver.Type]driver.SearchEngine) *Service {
	return &Service{
		drivers:    registered,
		selectOpts: make([]driver.SelectDriveOption, 0),
	}
}

func (s *Service) Search(ctx context.Context, request *pb.SearchRequest) (*pb.SearchResponse, error) {
	out := &pb.SearchResponse{}
	req := driver.SearchRequest{
		Source:    request.Source,
		Owner:     request.Owner,
		Query:     request.Query,
		Page:      request.Page,
		Condition: request.Condition,
	}

	// TODO: Multiple Driver Services One Response support.
	// assumption len(s.selectOpts) == 1.
	for i := range s.selectOpts {
		engine, ok := s.drivers[s.selectOpts[i]()]
		if !ok {
			return out, errors.New("no specified engine:" + string(s.selectOpts[i]()))
		}
		resp, err := engine.Search(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "build index error")
		}
		for j := range resp.Data {
			val, err := structpb.NewValue(resp.Data[j])
			if err != nil {
				return nil, errors.Wrap(err, "new value error")
			}
			out.Items = append(out.Items, val)
			out.Total += resp.Total
			out.Limit = resp.Limit
			out.Offset = resp.Offset
		}
	}

	return out, nil
}

func (s *Service) DeleteByID(ctx context.Context, request *pb.DeleteByIDRequest) (*pb.DeleteByIDResponse, error) {
	out := &pb.DeleteByIDResponse{}
	for i := range s.selectOpts {
		engine, ok := s.drivers[s.selectOpts[i]()]
		if !ok {
			return out, errors.New("no specified engine:" + string(s.selectOpts[i]()))
		}
		if err := engine.Delete(ctx, request.Id); err != nil {
			return out, errors.Wrap(err, "build index error")
		}
	}
	return out, nil
}

func (s *Service) Index(ctx context.Context, in *pb.IndexObject) (*pb.IndexResponse, error) {
	var (
		id  string
		out *pb.IndexResponse
	)
	out = &pb.IndexResponse{}

	switch kv := in.Obj.AsInterface().(type) {
	case map[string]interface{}:
		id = interface2string(kv["id"])
	case nil:
		out.Status = "SUCCESS"
		return out, nil
	default:
		return out, ErrIndexParamInvalid
	}
	objBytes, err := in.Obj.MarshalJSON()
	if err != nil {
		return out, errors.Wrap(err, "json marshal error")
	}
	for i := range s.selectOpts {
		engine, ok := s.drivers[s.selectOpts[i]()]
		if !ok {
			return out, errors.New("no specified engine:" + string(s.selectOpts[i]()))
		}
		if err = engine.BuildIndex(ctx, id, string(objBytes)); err != nil {
			return out, errors.Wrap(err, "build index error")
		}
	}
	out.Status = "SUCCESS"
	return out, nil
}

func (s *Service) SetSelectOptions(opts ...driver.SelectDriveOption) *Service {
	if len(opts) != 0 {
		s.selectOpts = opts
	}
	return s
}

func (s *Service) AppendSelectOptions(opts ...driver.SelectDriveOption) *Service {
	if len(opts) != 0 {
		s.selectOpts = append(s.selectOpts, opts...)
	}
	return s
}

func (s Service) Use(opts ...driver.SelectDriveOption) *Service {
	serv := s
	serv.selectOpts = nil
	if len(opts) != 0 {
		serv.selectOpts = opts
	}
	return &serv
}

func (s *Service) Register(name driver.Type, implement driver.SearchEngine) *Service {
	if s.drivers == nil {
		s.drivers = make(map[driver.Type]driver.SearchEngine)
	}
	s.drivers[name] = implement
	return s
}

func interface2string(in interface{}) (out string) {
	if in == nil {
		return
	}
	switch inString := in.(type) {
	case string:
		out = inString
	default:
		out = ""
	}
	return
}
