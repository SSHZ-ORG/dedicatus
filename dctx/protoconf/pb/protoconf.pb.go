// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: protoconf.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AuthConfig_UserType int32

const (
	AuthConfig_UNKNOWN_USER_TYPE AuthConfig_UserType = 0
	AuthConfig_USER              AuthConfig_UserType = 1 // Unused
	AuthConfig_CONTRIBUTOR       AuthConfig_UserType = 2
	AuthConfig_ADMIN             AuthConfig_UserType = 3
)

// Enum value maps for AuthConfig_UserType.
var (
	AuthConfig_UserType_name = map[int32]string{
		0: "UNKNOWN_USER_TYPE",
		1: "USER",
		2: "CONTRIBUTOR",
		3: "ADMIN",
	}
	AuthConfig_UserType_value = map[string]int32{
		"UNKNOWN_USER_TYPE": 0,
		"USER":              1,
		"CONTRIBUTOR":       2,
		"ADMIN":             3,
	}
)

func (x AuthConfig_UserType) Enum() *AuthConfig_UserType {
	p := new(AuthConfig_UserType)
	*p = x
	return p
}

func (x AuthConfig_UserType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AuthConfig_UserType) Descriptor() protoreflect.EnumDescriptor {
	return file_protoconf_proto_enumTypes[0].Descriptor()
}

func (AuthConfig_UserType) Type() protoreflect.EnumType {
	return &file_protoconf_proto_enumTypes[0]
}

func (x AuthConfig_UserType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *AuthConfig_UserType) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = AuthConfig_UserType(num)
	return nil
}

// Deprecated: Use AuthConfig_UserType.Descriptor instead.
func (AuthConfig_UserType) EnumDescriptor() ([]byte, []int) {
	return file_protoconf_proto_rawDescGZIP(), []int{1, 0}
}

type Protoconf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InlineQueryCacheTimeSec          *uint32     `protobuf:"varint,1,opt,name=inline_query_cache_time_sec,json=inlineQueryCacheTimeSec,def=0" json:"inline_query_cache_time_sec,omitempty"`
	TwapiLeastRecentPoolProbability  *float32    `protobuf:"fixed32,2,opt,name=twapi_least_recent_pool_probability,json=twapiLeastRecentPoolProbability,def=0.05" json:"twapi_least_recent_pool_probability,omitempty"`
	TwapiLeastRecentPoolOffsetRange  *uint32     `protobuf:"varint,3,opt,name=twapi_least_recent_pool_offset_range,json=twapiLeastRecentPoolOffsetRange,def=50" json:"twapi_least_recent_pool_offset_range,omitempty"`
	TwapiStandardPoolLimit           *uint32     `protobuf:"varint,4,opt,name=twapi_standard_pool_limit,json=twapiStandardPoolLimit,def=5" json:"twapi_standard_pool_limit,omitempty"`
	TwapiStandardPoolStepProbability *float32    `protobuf:"fixed32,5,opt,name=twapi_standard_pool_step_probability,json=twapiStandardPoolStepProbability,def=0.9" json:"twapi_standard_pool_step_probability,omitempty"`
	AuthConfig                       *AuthConfig `protobuf:"bytes,100,opt,name=auth_config,json=authConfig" json:"auth_config,omitempty"`
}

// Default values for Protoconf fields.
const (
	Default_Protoconf_InlineQueryCacheTimeSec          = uint32(0)
	Default_Protoconf_TwapiLeastRecentPoolProbability  = float32(0.05000000074505806)
	Default_Protoconf_TwapiLeastRecentPoolOffsetRange  = uint32(50)
	Default_Protoconf_TwapiStandardPoolLimit           = uint32(5)
	Default_Protoconf_TwapiStandardPoolStepProbability = float32(0.8999999761581421)
)

func (x *Protoconf) Reset() {
	*x = Protoconf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoconf_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Protoconf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Protoconf) ProtoMessage() {}

func (x *Protoconf) ProtoReflect() protoreflect.Message {
	mi := &file_protoconf_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Protoconf.ProtoReflect.Descriptor instead.
func (*Protoconf) Descriptor() ([]byte, []int) {
	return file_protoconf_proto_rawDescGZIP(), []int{0}
}

func (x *Protoconf) GetInlineQueryCacheTimeSec() uint32 {
	if x != nil && x.InlineQueryCacheTimeSec != nil {
		return *x.InlineQueryCacheTimeSec
	}
	return Default_Protoconf_InlineQueryCacheTimeSec
}

func (x *Protoconf) GetTwapiLeastRecentPoolProbability() float32 {
	if x != nil && x.TwapiLeastRecentPoolProbability != nil {
		return *x.TwapiLeastRecentPoolProbability
	}
	return Default_Protoconf_TwapiLeastRecentPoolProbability
}

func (x *Protoconf) GetTwapiLeastRecentPoolOffsetRange() uint32 {
	if x != nil && x.TwapiLeastRecentPoolOffsetRange != nil {
		return *x.TwapiLeastRecentPoolOffsetRange
	}
	return Default_Protoconf_TwapiLeastRecentPoolOffsetRange
}

func (x *Protoconf) GetTwapiStandardPoolLimit() uint32 {
	if x != nil && x.TwapiStandardPoolLimit != nil {
		return *x.TwapiStandardPoolLimit
	}
	return Default_Protoconf_TwapiStandardPoolLimit
}

func (x *Protoconf) GetTwapiStandardPoolStepProbability() float32 {
	if x != nil && x.TwapiStandardPoolStepProbability != nil {
		return *x.TwapiStandardPoolStepProbability
	}
	return Default_Protoconf_TwapiStandardPoolStepProbability
}

func (x *Protoconf) GetAuthConfig() *AuthConfig {
	if x != nil {
		return x.AuthConfig
	}
	return nil
}

type AuthConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Users map[int64]AuthConfig_UserType `protobuf:"bytes,1,rep,name=users" json:"users,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value,enum=dedicatus.dctx.protoconf.pb.AuthConfig_UserType"`
}

func (x *AuthConfig) Reset() {
	*x = AuthConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoconf_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthConfig) ProtoMessage() {}

func (x *AuthConfig) ProtoReflect() protoreflect.Message {
	mi := &file_protoconf_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthConfig.ProtoReflect.Descriptor instead.
func (*AuthConfig) Descriptor() ([]byte, []int) {
	return file_protoconf_proto_rawDescGZIP(), []int{1}
}

func (x *AuthConfig) GetUsers() map[int64]AuthConfig_UserType {
	if x != nil {
		return x.Users
	}
	return nil
}

var File_protoconf_proto protoreflect.FileDescriptor

var file_protoconf_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6e, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x1b, 0x64, 0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x64, 0x63, 0x74,
	0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6e, 0x66, 0x2e, 0x70, 0x62, 0x22, 0xd0,
	0x03, 0x0a, 0x09, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6e, 0x66, 0x12, 0x3f, 0x0a, 0x1b,
	0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x63, 0x61, 0x63,
	0x68, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x73, 0x65, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x3a, 0x01, 0x30, 0x52, 0x17, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x43, 0x61, 0x63, 0x68, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x53, 0x65, 0x63, 0x12, 0x52, 0x0a,
	0x23, 0x74, 0x77, 0x61, 0x70, 0x69, 0x5f, 0x6c, 0x65, 0x61, 0x73, 0x74, 0x5f, 0x72, 0x65, 0x63,
	0x65, 0x6e, 0x74, 0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x5f, 0x70, 0x72, 0x6f, 0x62, 0x61, 0x62, 0x69,
	0x6c, 0x69, 0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x3a, 0x04, 0x30, 0x2e, 0x30, 0x35,
	0x52, 0x1f, 0x74, 0x77, 0x61, 0x70, 0x69, 0x4c, 0x65, 0x61, 0x73, 0x74, 0x52, 0x65, 0x63, 0x65,
	0x6e, 0x74, 0x50, 0x6f, 0x6f, 0x6c, 0x50, 0x72, 0x6f, 0x62, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x79, 0x12, 0x51, 0x0a, 0x24, 0x74, 0x77, 0x61, 0x70, 0x69, 0x5f, 0x6c, 0x65, 0x61, 0x73, 0x74,
	0x5f, 0x72, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x5f, 0x6f, 0x66, 0x66,
	0x73, 0x65, 0x74, 0x5f, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x3a,
	0x02, 0x35, 0x30, 0x52, 0x1f, 0x74, 0x77, 0x61, 0x70, 0x69, 0x4c, 0x65, 0x61, 0x73, 0x74, 0x52,
	0x65, 0x63, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x6f, 0x6c, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x52,
	0x61, 0x6e, 0x67, 0x65, 0x12, 0x3c, 0x0a, 0x19, 0x74, 0x77, 0x61, 0x70, 0x69, 0x5f, 0x73, 0x74,
	0x61, 0x6e, 0x64, 0x61, 0x72, 0x64, 0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x5f, 0x6c, 0x69, 0x6d, 0x69,
	0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x3a, 0x01, 0x35, 0x52, 0x16, 0x74, 0x77, 0x61, 0x70,
	0x69, 0x53, 0x74, 0x61, 0x6e, 0x64, 0x61, 0x72, 0x64, 0x50, 0x6f, 0x6f, 0x6c, 0x4c, 0x69, 0x6d,
	0x69, 0x74, 0x12, 0x53, 0x0a, 0x24, 0x74, 0x77, 0x61, 0x70, 0x69, 0x5f, 0x73, 0x74, 0x61, 0x6e,
	0x64, 0x61, 0x72, 0x64, 0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x5f, 0x73, 0x74, 0x65, 0x70, 0x5f, 0x70,
	0x72, 0x6f, 0x62, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02,
	0x3a, 0x03, 0x30, 0x2e, 0x39, 0x52, 0x20, 0x74, 0x77, 0x61, 0x70, 0x69, 0x53, 0x74, 0x61, 0x6e,
	0x64, 0x61, 0x72, 0x64, 0x50, 0x6f, 0x6f, 0x6c, 0x53, 0x74, 0x65, 0x70, 0x50, 0x72, 0x6f, 0x62,
	0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x48, 0x0a, 0x0b, 0x61, 0x75, 0x74, 0x68, 0x5f,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x64,
	0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x64, 0x63, 0x74, 0x78, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6e, 0x66, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0a, 0x61, 0x75, 0x74, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x22, 0x8b, 0x02, 0x0a, 0x0a, 0x41, 0x75, 0x74, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x48, 0x0a, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x32, 0x2e, 0x64, 0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x64, 0x63, 0x74, 0x78,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6e, 0x66, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x75,
	0x74, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x1a, 0x6a, 0x0a, 0x0a, 0x55, 0x73,
	0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x46, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x30, 0x2e, 0x64, 0x65, 0x64, 0x69,
	0x63, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x64, 0x63, 0x74, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6e, 0x66, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x47, 0x0a, 0x08, 0x55, 0x73, 0x65, 0x72, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x55, 0x53,
	0x45, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x55, 0x53, 0x45,
	0x52, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x4f, 0x4e, 0x54, 0x52, 0x49, 0x42, 0x55, 0x54,
	0x4f, 0x52, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x41, 0x44, 0x4d, 0x49, 0x4e, 0x10, 0x03, 0x42,
	0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x53,
	0x48, 0x5a, 0x2d, 0x4f, 0x52, 0x47, 0x2f, 0x64, 0x65, 0x64, 0x69, 0x63, 0x61, 0x74, 0x75, 0x73,
	0x2f, 0x64, 0x63, 0x74, 0x78, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6e, 0x66, 0x2f,
	0x70, 0x62,
}

var (
	file_protoconf_proto_rawDescOnce sync.Once
	file_protoconf_proto_rawDescData = file_protoconf_proto_rawDesc
)

func file_protoconf_proto_rawDescGZIP() []byte {
	file_protoconf_proto_rawDescOnce.Do(func() {
		file_protoconf_proto_rawDescData = protoimpl.X.CompressGZIP(file_protoconf_proto_rawDescData)
	})
	return file_protoconf_proto_rawDescData
}

var file_protoconf_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protoconf_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protoconf_proto_goTypes = []interface{}{
	(AuthConfig_UserType)(0), // 0: dedicatus.dctx.protoconf.pb.AuthConfig.UserType
	(*Protoconf)(nil),        // 1: dedicatus.dctx.protoconf.pb.Protoconf
	(*AuthConfig)(nil),       // 2: dedicatus.dctx.protoconf.pb.AuthConfig
	nil,                      // 3: dedicatus.dctx.protoconf.pb.AuthConfig.UsersEntry
}
var file_protoconf_proto_depIdxs = []int32{
	2, // 0: dedicatus.dctx.protoconf.pb.Protoconf.auth_config:type_name -> dedicatus.dctx.protoconf.pb.AuthConfig
	3, // 1: dedicatus.dctx.protoconf.pb.AuthConfig.users:type_name -> dedicatus.dctx.protoconf.pb.AuthConfig.UsersEntry
	0, // 2: dedicatus.dctx.protoconf.pb.AuthConfig.UsersEntry.value:type_name -> dedicatus.dctx.protoconf.pb.AuthConfig.UserType
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_protoconf_proto_init() }
func file_protoconf_proto_init() {
	if File_protoconf_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protoconf_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Protoconf); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protoconf_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protoconf_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protoconf_proto_goTypes,
		DependencyIndexes: file_protoconf_proto_depIdxs,
		EnumInfos:         file_protoconf_proto_enumTypes,
		MessageInfos:      file_protoconf_proto_msgTypes,
	}.Build()
	File_protoconf_proto = out.File
	file_protoconf_proto_rawDesc = nil
	file_protoconf_proto_goTypes = nil
	file_protoconf_proto_depIdxs = nil
}
