package service

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type RawdataService struct {
	pb.UnimplementedRawdataServer
	rawdataClient rawdata.RawDataService
}

func NewRawdataService() (*RawdataService, error) {
	rawdataClient := rawdata.NewRawDataService(config.Get().Components.RawData.Name)
	if err := rawdataClient.Init(resource.ParseFrom(config.Get().Components.RawData)); err != nil {
		log.L().Error("initialize rawdata server", zap.Error(err))
	}
	return &RawdataService{
		rawdataClient: rawdataClient,
	}, nil
}

func (s *RawdataService) GetRawdata(ctx context.Context, req *pb.GetRawdataRequest) (*pb.GetRawdataResponse, error) {
	if req.StartTime < (time.Now().Unix() - 3600*24*3) {
		req.StartTime = time.Now().Unix() - 3600*24*3
	}

	if err := checkParams(req.StartTime, req.EndTime, req.Path); err != nil {
		return nil, err
	}

	user := defalutUser
	h := ctx.Value(contextHTTPHeaderKey)
	header, ok := h.(http.Header)
	if ok {
		user = getUser(header, user)
	}
	// 检查user和实体id的合法性

	resp, err := s.rawdataClient.Query(ctx, req)

	return resp, err
}

func getUser(header http.Header, oldUser string) string {
	auth := header.Get("X-Tkeel-Auth")
	log.L().Info("user: ", zap.String("auth", auth))
	if bytes, err := base64.StdEncoding.DecodeString(auth); err == nil {
		urlquery, err1 := url.ParseQuery(string(bytes))
		if err1 == nil {
			return urlquery.Get("user")
		}
	} else {
		log.L().Error("auth error", zap.Error(err))
	}
	return oldUser
}
