/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

//func TestDao_DelCostumeResource(t *testing.T) {
//	type fields struct {
//		ctx          context.Context
//		cancel       context.CancelFunc
//		storeCfg     config.Metadata
//		etcdCfg      config.EtcdConfig
//		stateClient  store.Store
//		etcdEndpoint *clientv3.Client
//		entityCodec  entityCodec
//	}
//	type args struct {
//		ctx  context.Context
//		expr CostumeResource
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Dao{
//				ctx:          tt.fields.ctx,
//				cancel:       tt.fields.cancel,
//				storeCfg:     tt.fields.storeCfg,
//				etcdCfg:      tt.fields.etcdCfg,
//				stateClient:  tt.fields.stateClient,
//				etcdEndpoint: tt.fields.etcdEndpoint,
//				entityCodec:  tt.fields.entityCodec,
//			}
//			if err := d.DelCostumeResource(tt.args.ctx, tt.args.expr); (err != nil) != tt.wantErr {
//				t.Errorf("DelCostumeResource() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestDao_GetCostumeResource(t *testing.T) {
//	type fields struct {
//		ctx          context.Context
//		cancel       context.CancelFunc
//		storeCfg     config.Metadata
//		etcdCfg      config.EtcdConfig
//		stateClient  store.Store
//		etcdEndpoint *clientv3.Client
//		entityCodec  entityCodec
//	}
//	type args struct {
//		ctx  context.Context
//		expr CostumeResource
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    CostumeResource
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Dao{
//				ctx:          tt.fields.ctx,
//				cancel:       tt.fields.cancel,
//				storeCfg:     tt.fields.storeCfg,
//				etcdCfg:      tt.fields.etcdCfg,
//				stateClient:  tt.fields.stateClient,
//				etcdEndpoint: tt.fields.etcdEndpoint,
//				entityCodec:  tt.fields.entityCodec,
//			}
//			got, err := d.GetCostumeResource(tt.args.ctx, tt.args.expr)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetCostumeResource() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetCostumeResource() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestDao_HasCostumeResource(t *testing.T) {
//	type fields struct {
//		ctx          context.Context
//		cancel       context.CancelFunc
//		storeCfg     config.Metadata
//		etcdCfg      config.EtcdConfig
//		stateClient  store.Store
//		etcdEndpoint *clientv3.Client
//		entityCodec  entityCodec
//	}
//	type args struct {
//		ctx  context.Context
//		expr CostumeResource
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    bool
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Dao{
//				ctx:          tt.fields.ctx,
//				cancel:       tt.fields.cancel,
//				storeCfg:     tt.fields.storeCfg,
//				etcdCfg:      tt.fields.etcdCfg,
//				stateClient:  tt.fields.stateClient,
//				etcdEndpoint: tt.fields.etcdEndpoint,
//				entityCodec:  tt.fields.entityCodec,
//			}
//			got, err := d.HasCostumeResource(tt.args.ctx, tt.args.expr)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("HasCostumeResource() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("HasCostumeResource() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestDao_ListCostumeResource(t *testing.T) {
//	type fields struct {
//		ctx          context.Context
//		cancel       context.CancelFunc
//		storeCfg     config.Metadata
//		etcdCfg      config.EtcdConfig
//		stateClient  store.Store
//		etcdEndpoint *clientv3.Client
//		entityCodec  entityCodec
//	}
//	type args struct {
//		ctx context.Context
//		rev int64
//		req CostumeResourceReq
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    []CostumeResource
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Dao{
//				ctx:          tt.fields.ctx,
//				cancel:       tt.fields.cancel,
//				storeCfg:     tt.fields.storeCfg,
//				etcdCfg:      tt.fields.etcdCfg,
//				stateClient:  tt.fields.stateClient,
//				etcdEndpoint: tt.fields.etcdEndpoint,
//				entityCodec:  tt.fields.entityCodec,
//			}
//			got, err := d.ListCostumeResource(tt.args.ctx, tt.args.rev, tt.args.req)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ListCostumeResource() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("ListCostumeResource() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestDao_PutCostumeResource(t *testing.T) {
//	type fields struct {
//		ctx          context.Context
//		cancel       context.CancelFunc
//		storeCfg     config.Metadata
//		etcdCfg      config.EtcdConfig
//		stateClient  store.Store
//		etcdEndpoint *clientv3.Client
//		entityCodec  entityCodec
//	}
//	type args struct {
//		ctx  context.Context
//		expr CostumeResource
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Dao{
//				ctx:          tt.fields.ctx,
//				cancel:       tt.fields.cancel,
//				storeCfg:     tt.fields.storeCfg,
//				etcdCfg:      tt.fields.etcdCfg,
//				stateClient:  tt.fields.stateClient,
//				etcdEndpoint: tt.fields.etcdEndpoint,
//				entityCodec:  tt.fields.entityCodec,
//			}
//			if err := d.PutCostumeResource(tt.args.ctx, tt.args.expr); (err != nil) != tt.wantErr {
//				t.Errorf("PutCostumeResource() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestDao_RangeCostumeResource(t *testing.T) {
//	type fields struct {
//		ctx          context.Context
//		cancel       context.CancelFunc
//		storeCfg     config.Metadata
//		etcdCfg      config.EtcdConfig
//		stateClient  store.Store
//		etcdEndpoint *clientv3.Client
//		entityCodec  entityCodec
//	}
//	type args struct {
//		ctx     context.Context
//		rev     int64
//		handler ListCostumeResourceFunc
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Dao{
//				ctx:          tt.fields.ctx,
//				cancel:       tt.fields.cancel,
//				storeCfg:     tt.fields.storeCfg,
//				etcdCfg:      tt.fields.etcdCfg,
//				stateClient:  tt.fields.stateClient,
//				etcdEndpoint: tt.fields.etcdEndpoint,
//				entityCodec:  tt.fields.entityCodec,
//			}
//		})
//	}
//}
//
//func TestDao_WatchCostumeResource(t *testing.T) {
//	type fields struct {
//		ctx          context.Context
//		cancel       context.CancelFunc
//		storeCfg     config.Metadata
//		etcdCfg      config.EtcdConfig
//		stateClient  store.Store
//		etcdEndpoint *clientv3.Client
//		entityCodec  entityCodec
//	}
//	type args struct {
//		ctx     context.Context
//		rev     int64
//		handler WatchCostumeResourceFunc
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Dao{
//				ctx:          tt.fields.ctx,
//				cancel:       tt.fields.cancel,
//				storeCfg:     tt.fields.storeCfg,
//				etcdCfg:      tt.fields.etcdCfg,
//				stateClient:  tt.fields.stateClient,
//				etcdEndpoint: tt.fields.etcdEndpoint,
//				entityCodec:  tt.fields.entityCodec,
//			}
//		})
//	}
//}
