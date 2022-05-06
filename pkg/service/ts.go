package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	logf "github.com/tkeel-io/core/pkg/logfield"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/tkeel-io/kit/log"
)

var defalutUser = "admin"

type TSService struct {
	pb.UnimplementedTSServer
	tseriesClient tseries.TimeSerier
	entityCache   map[string][]string
	apiManager    apim.APIManager
	lock          *sync.RWMutex
	inited        *atomic.Bool
}

func NewTSService() (*TSService, error) {
	tseriesClient := tseries.NewTimeSerier(config.Get().Components.TimeSeries.Name)
	if err := tseriesClient.Init(resource.ParseFrom(config.Get().Components.TimeSeries)); nil != err {
		log.L().Error("initialize time series", logf.Error(err))
		return nil, errors.Wrap(err, "init ts service")
	}

	return &TSService{
		tseriesClient: tseriesClient,
		entityCache:   make(map[string][]string),
		lock:          new(sync.RWMutex),
		inited:        atomic.NewBool(false),
	}, nil
}

func (s *TSService) Init(apiManager apim.APIManager) {
	s.inited.Store(true)
	s.apiManager = apiManager
}

func (s *TSService) AddEntity(user, entityID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.entityCache[user]; !ok {
		s.entityCache[user] = make([]string, 0)
	}
	length := len(s.entityCache[user])
	if length == 0 {
		s.entityCache[user] = append(s.entityCache[user], entityID)
		return
	}
	oldCache := s.entityCache[user]
	cacheMap := make(map[string]struct{})
	s.entityCache[user] = make([]string, 0)
	s.entityCache[user] = append(s.entityCache[user], entityID)
	cacheMap[entityID] = struct{}{}
	count := 1
	for _, v := range oldCache {
		if _, ok := cacheMap[v]; ok {
			continue
		}
		s.entityCache[user] = append(s.entityCache[user], v)
		cacheMap[v] = struct{}{}
		count++
		if count >= 5 {
			break
		}
	}
}

func (s *TSService) GetTSData(ctx context.Context, req *pb.GetTSDataRequest) (*pb.GetTSDataResponse, error) {
	if req.StartTime < (time.Now().Unix() - 3600*24*3) {
		req.StartTime = time.Now().Unix() - 3600*24*3
	}
	if err := checkParams(req.StartTime, req.EndTime, req.Identifiers); err != nil {
		return nil, err
	}

	user := defalutUser
	h := ctx.Value(contextHTTPHeaderKey)
	header, ok := h.(http.Header)
	if ok {
		auth := header.Get("X-Tkeel-Auth")
		log.L().Info("user: ", logf.String("auth", auth))
		if bytes, err := base64.StdEncoding.DecodeString(auth); err == nil {
			urlquery, err1 := url.ParseQuery(string(bytes))
			if err1 == nil {
				user = urlquery.Get("user")
			}
		} else {
			log.L().Error("auth error", logf.Error(err))
		}
	}
	resp := &pb.GetTSDataResponse{}
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 0
	}

	res, err := s.tseriesClient.Query(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "query time series data")
	}
	resp.Total = res.Total
	resp.Items = res.Items
	if resp.Total > 0 {
		s.AddEntity(user, req.Id)
	}
	return resp, nil
}

func checkParams(startTime, endTime int64, identifiers string) error {
	if startTime < (time.Now().Unix() - 3600*24*3 - 3600) {
		return errors.New("time error")
	} else if startTime > endTime {
		return errors.New("time error")
	} else if identifiers == "" {
		return errors.New("identifiers error")
	}
	return nil
}

func (s *TSService) DownloadTSData(ctx context.Context, req *pb.DownloadTSDataRequest) (*pb.DownloadTSDataResponse, error) {
	resp := &pb.DownloadTSDataResponse{}
	if req.StartTime < (time.Now().Unix() - 3600*24*3) {
		req.StartTime = time.Now().Unix() - 3600*24*3
	}

	if err := checkParams(req.StartTime, req.EndTime, req.Identifiers); err != nil {
		resp.Data = []byte("error")
		resp.Length = "5"
		resp.Filename = "error.txt"
		return resp, err
	}

	var buffer []byte
	csvBuffer := bytes.NewBuffer(buffer)
	csvWriter := csv.NewWriter(csvBuffer)
	base := []string{"id", "time"}
	identifiers := strings.Split(req.Identifiers, ",")
	base = append(base, identifiers...)
	csvWriter.Write(base)
	var pageSize int32 = 300

	run := true
	for i := 1; run; i++ {
		reqGet := &pb.GetTSDataRequest{
			Id:          req.Id,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Identifiers: req.Identifiers,
			PageNum:     int32(i),
			PageSize:    pageSize,
		}

		res, err := s.tseriesClient.Query(ctx, reqGet)
		if err != nil {
			resp.Data = []byte("error")
			resp.Length = "5"
			resp.Filename = "error.txt"
			return resp, errors.Wrap(err, "query time series data")
		}
		if res.Total < pageSize {
			run = false
		}
		for _, v := range res.Items {
			base := []string{req.Id, fmt.Sprintf("%d", v.Time)}
			for _, identifier := range identifiers {
				if vv, ok := v.Value[identifier]; ok {
					base = append(base, fmt.Sprintf("%f", vv))
				} else {
					base = append(base, "")
				}
			}
			csvWriter.Write(base)
		}
	}
	csvWriter.Flush()
	resp.Data = csvBuffer.Bytes()

	resp.Length = fmt.Sprintf("%d", len(resp.Data))
	resp.Filename = fmt.Sprintf("%s_%s.csv", req.Id, time.Now().Format("2006-01-02-15-04-05"))

	return resp, nil
}

var contextHTTPHeaderKey = struct{}{}

func Entity2EntityResponse(base *apim.BaseRet) (out *pb.EntityResponse, err error) {
	if base == nil {
		return
	}

	out = &pb.EntityResponse{}
	if out.Properties, err = structpb.NewValue(base.Properties); nil != err {
		log.L().Error("convert entity properties", logf.Error(err), logf.ID(base.ID))
		return out, errors.Wrap(err, "convert entity properties")
	} else if out.Configs, err = structpb.NewValue(base.Scheme); nil != err {
		log.L().Error("convert entity scheme.", logf.Error(err), logf.ID(base.ID))
		return out, errors.Wrap(err, "convert entity scheme")
	}

	out.Mappers = make([]*pb.Mapper, 0)
	for _, mDesc := range base.Mappers {
		out.Mappers = append(out.Mappers,
			&pb.Mapper{
				Id:          mDesc.Id,
				Name:        mDesc.Name,
				Tql:         mDesc.Tql,
				Description: mDesc.Description,
			})
	}

	out.Id = base.ID
	out.Type = base.Type
	out.Owner = base.Owner
	out.Source = base.Source
	out.Version = base.Version
	out.LastTime = base.LastTime
	return out, nil
}

func (s *TSService) GetLatestEntities(ctx context.Context, req *pb.GetLatestEntitiesRequest) (resp *pb.GetLatestEntitiesResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready")
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	resp = &pb.GetLatestEntitiesResponse{}
	user := defalutUser
	h := ctx.Value(contextHTTPHeaderKey)
	header, ok := h.(http.Header)
	if ok {
		var bytes []byte
		auth := header.Get("X-Tkeel-Auth")
		log.Info("user: ", auth)
		if bytes, err = base64.StdEncoding.DecodeString(auth); err == nil {
			urlquery, err1 := url.ParseQuery(string(bytes))
			if err1 == nil {
				user = urlquery.Get("user")
			}
		} else {
			log.Error(err)
		}
	} else {
		return
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	if cache, ok := s.entityCache[user]; !ok {
		resp.Total = 0
		resp.Items = make([]*pb.EntityResponse, 0)
		return resp, nil
	} else { //nolint
		resp.Total = int64(len(cache))
		for _, v := range cache {
			var entityBase = new(Entity)
			entityBase.ID = v
			entityBase.Source = "source"
			entityBase.Owner = user

			var baseRet *apim.BaseRet
			if baseRet, err = s.apiManager.GetEntity(ctx, entityBase); nil != err {
				log.L().Error("get entity", logf.Eid(entityBase.ID), logf.Error(err))
				continue
			}

			var entity *pb.EntityResponse
			if entity, err = Entity2EntityResponse(baseRet); nil != err {
				log.Error("patch entity failed.", logf.Eid(v), logf.Error(err))
			} else {
				resp.Items = append(resp.Items, entity)
			}
		}
	}
	return resp, nil
}
