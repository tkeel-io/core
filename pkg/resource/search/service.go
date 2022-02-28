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

func Init() *Service {
	defaultRegistered := map[driver.Type]driver.SearchEngine{
		// Add other drivers to SearchService here.
		driver.ElasticsearchDriver: driver.NewElasticsearchEngine(config.Get().Components.SearchEngine.ES),
	}
	return NewService(defaultRegistered).Use(driver.Parse(config.Get().Components.SearchEngine.Use))
}

var _ pb.SearchHTTPServer = &Service{}

type Service struct {
	drivers   map[driver.Type]driver.SearchEngine
	selectOpt driver.SelectDriveOption
}

func NewService(registered map[driver.Type]driver.SearchEngine) *Service {
	return &Service{
		drivers:   registered,
		selectOpt: driver.NoopDriver,
	}
}

func (s *Service) Search(ctx context.Context, request *pb.SearchRequest) (*pb.SearchResponse, error) {
	out := &pb.SearchResponse{}
	req := driver.SearchRequest{
		Source:    request.Source,
		Owner:     request.Owner,
		Query:     request.Query,
		Condition: request.Condition,
	}
	req.Page = &pb.Pager{}
	req.Page.Limit = request.PageSize
	req.Page.Offset = request.PageSize * (request.PageNum - 1)
	req.Page.Reverse = request.IsDescending
	req.Page.Sort = request.OrderBy

	// TODO: Multiple Driver Services One Response support.
	// assumption len(s.selectOpt) == 1.

	engine, ok := s.drivers[s.selectOpt()]
	if !ok {
		return out, errors.New("no specified engine:" + string(s.selectOpt()))
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
	}
	out.Total = resp.Total
	out.PageNum = request.PageNum
	out.PageSize = request.PageSize

	return out, nil
}

func (s *Service) DeleteByID(ctx context.Context, request *pb.DeleteByIDRequest) (*pb.DeleteByIDResponse, error) {
	out := &pb.DeleteByIDResponse{}
	engine, ok := s.drivers[s.selectOpt()]
	if !ok {
		return out, errors.New("no specified engine:" + string(s.selectOpt()))
	}
	if err := engine.Delete(ctx, request.Id); err != nil {
		return out, errors.Wrap(err, "build index error")
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
	engine, ok := s.drivers[s.selectOpt()]
	if !ok {
		return out, errors.New("no specified engine:" + string(s.selectOpt()))
	}
	if err = engine.BuildIndex(ctx, id, string(objBytes)); err != nil {
		return out, errors.Wrap(err, "build index error")
	}
	out.Status = "SUCCESS"
	return out, nil
}

// Use SelectDriveOption and set the option to this service.
func (s *Service) Use(opt driver.SelectDriveOption) *Service {
	s.selectOpt = opt
	return s
}

// With SelectDriveOption create a copy from original service.
func (s Service) With(opt driver.SelectDriveOption) *Service {
	serv := s
	serv.selectOpt = opt
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
