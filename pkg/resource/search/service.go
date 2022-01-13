package search

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/resource/search/driver"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

var Service = newService()

type service struct {
	drivers    map[driver.Type]driver.Engine
	driverOpts []driver.SelectDriveOption
}

func newService() *service {
	return &service{drivers: map[driver.Type]driver.Engine{
		driver.Elasticsearch: driver.NewElasticsearchEngine(config.Get().SearchEngine.ES.Urls...),
	}}
}

func (s *service) Search(ctx context.Context, request *pb.SearchRequest) (*pb.SearchResponse, error) {
	out := &pb.SearchResponse{}
	req := driver.SearchRequest{
		Source:    request.Source,
		Owner:     request.Owner,
		Query:     request.Query,
		Page:      request.Page,
		Condition: request.Condition,
	}

	// TODO: Multiple Driver Services One Response support.
	// assumption len(s.driverOpts) == 1.
	for i := range s.driverOpts {
		resp, err := s.drivers[s.driverOpts[i]()].Search(ctx, req)
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

func (s *service) DeleteByID(ctx context.Context, request *pb.DeleteByIDRequest) (*pb.DeleteByIDResponse, error) {
	out := &pb.DeleteByIDResponse{}
	for i := range s.driverOpts {
		if err := s.drivers[s.driverOpts[i]()].Delete(ctx, request.Id); err != nil {
			return out, errors.Wrap(err, "build index error")
		}
	}
	return out, nil
}

func (s *service) Index(ctx context.Context, in *pb.IndexObject) (*pb.IndexResponse, error) {
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
	for i := range s.driverOpts {
		if err = s.drivers[s.driverOpts[i]()].BuildIndex(ctx, id, string(objBytes)); err != nil {
			return out, errors.Wrap(err, "build index error")
		}
	}
	out.Status = "SUCCESS"
	return out, nil
}

func (s *service) SelectDrive(opts ...driver.SelectDriveOption) *service {
	if len(opts) != 0 {
		s.driverOpts = opts
	}
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
