// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.15.8
// source: api/openapi/v1/server.proto

package v1

import (
	v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_api_openapi_v1_server_proto protoreflect.FileDescriptor

var file_api_openapi_v1_server_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x61, 0x70, 0x69, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6f,
	0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x1a, 0x18, 0x6f, 0x70, 0x65, 0x6e, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xf4,
	0x03, 0x0a, 0x07, 0x4f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x12, 0x53, 0x0a, 0x08, 0x49, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x66, 0x79, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1c,
	0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x11, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x0b, 0x12, 0x09, 0x2f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x79, 0x12,
	0x74, 0x0a, 0x0e, 0x41, 0x64, 0x64, 0x6f, 0x6e, 0x73, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66,
	0x79, 0x12, 0x21, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x41,
	0x64, 0x64, 0x6f, 0x6e, 0x73, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x41, 0x64, 0x64, 0x6f, 0x6e, 0x73, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x79,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15,
	0x22, 0x10, 0x2f, 0x61, 0x64, 0x64, 0x6f, 0x6e, 0x73, 0x2f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x66, 0x79, 0x3a, 0x01, 0x2a, 0x12, 0x63, 0x0a, 0x0a, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x42,
	0x69, 0x6e, 0x64, 0x12, 0x1c, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x42, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x73,
	0x74, 0x1a, 0x1e, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x54,
	0x65, 0x6e, 0x61, 0x6e, 0x74, 0x42, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x17, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11, 0x22, 0x0c, 0x2f, 0x74, 0x65, 0x6e, 0x61,
	0x6e, 0x74, 0x2f, 0x62, 0x69, 0x6e, 0x64, 0x3a, 0x01, 0x2a, 0x12, 0x6b, 0x0a, 0x0c, 0x54, 0x65,
	0x6e, 0x61, 0x6e, 0x74, 0x55, 0x6e, 0x62, 0x69, 0x6e, 0x64, 0x12, 0x1e, 0x2e, 0x6f, 0x70, 0x65,
	0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x55, 0x6e,
	0x62, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x6f, 0x70, 0x65,
	0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x55, 0x6e,
	0x62, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x19, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x13, 0x22, 0x0e, 0x2f, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x2f, 0x75, 0x6e,
	0x62, 0x69, 0x6e, 0x64, 0x3a, 0x01, 0x2a, 0x12, 0x4c, 0x0a, 0x05, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1a, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x0f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x09, 0x12, 0x07, 0x2f, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x5e, 0x0a, 0x1e, 0x64, 0x65, 0x76, 0x2e, 0x74, 0x6b, 0x65,
	0x65, 0x6c, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x6f, 0x70, 0x65,
	0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x42, 0x0e, 0x4f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x56, 0x31, 0x50, 0x01, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x6b, 0x65, 0x65, 0x6c, 0x2d, 0x69, 0x6f, 0x2f, 0x63,
	0x6f, 0x72, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_api_openapi_v1_server_proto_goTypes = []interface{}{
	(*emptypb.Empty)(nil),             // 0: google.protobuf.Empty
	(*v1.AddonsIdentifyRequest)(nil),  // 1: openapi.v1.AddonsIdentifyRequest
	(*v1.TenantBindRequst)(nil),       // 2: openapi.v1.TenantBindRequst
	(*v1.TenantUnbindRequst)(nil),     // 3: openapi.v1.TenantUnbindRequst
	(*v1.IdentifyResponse)(nil),       // 4: openapi.v1.IdentifyResponse
	(*v1.AddonsIdentifyResponse)(nil), // 5: openapi.v1.AddonsIdentifyResponse
	(*v1.TenantBindResponse)(nil),     // 6: openapi.v1.TenantBindResponse
	(*v1.TenantUnbindResponse)(nil),   // 7: openapi.v1.TenantUnbindResponse
	(*v1.StatusResponse)(nil),         // 8: openapi.v1.StatusResponse
}
var file_api_openapi_v1_server_proto_depIdxs = []int32{
	0, // 0: openapi.v1.Openapi.Identify:input_type -> google.protobuf.Empty
	1, // 1: openapi.v1.Openapi.AddonsIdentify:input_type -> openapi.v1.AddonsIdentifyRequest
	2, // 2: openapi.v1.Openapi.TenantBind:input_type -> openapi.v1.TenantBindRequst
	3, // 3: openapi.v1.Openapi.TenantUnbind:input_type -> openapi.v1.TenantUnbindRequst
	0, // 4: openapi.v1.Openapi.tatus:input_type -> google.protobuf.Empty
	4, // 5: openapi.v1.Openapi.Identify:output_type -> openapi.v1.IdentifyResponse
	5, // 6: openapi.v1.Openapi.AddonsIdentify:output_type -> openapi.v1.AddonsIdentifyResponse
	6, // 7: openapi.v1.Openapi.TenantBind:output_type -> openapi.v1.TenantBindResponse
	7, // 8: openapi.v1.Openapi.TenantUnbind:output_type -> openapi.v1.TenantUnbindResponse
	8, // 9: openapi.v1.Openapi.tatus:output_type -> openapi.v1.StatusResponse
	5, // [5:10] is the sub-list for method output_type
	0, // [0:5] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_openapi_v1_server_proto_init() }
func file_api_openapi_v1_server_proto_init() {
	if File_api_openapi_v1_server_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_openapi_v1_server_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_openapi_v1_server_proto_goTypes,
		DependencyIndexes: file_api_openapi_v1_server_proto_depIdxs,
	}.Build()
	File_api_openapi_v1_server_proto = out.File
	file_api_openapi_v1_server_proto_rawDesc = nil
	file_api_openapi_v1_server_proto_goTypes = nil
	file_api_openapi_v1_server_proto_depIdxs = nil
}
