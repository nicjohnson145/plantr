// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: plantr/controller/v1/struct.proto

package controllerv1

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

type VersionType int32

const (
	VersionType_VERSION_TYPE_UNSPECIFIED VersionType = 0
	VersionType_VERSION_TYPE_PINNED      VersionType = 1
	VersionType_VERSION_TYPE_LATEST      VersionType = 2
)

// Enum value maps for VersionType.
var (
	VersionType_name = map[int32]string{
		0: "VERSION_TYPE_UNSPECIFIED",
		1: "VERSION_TYPE_PINNED",
		2: "VERSION_TYPE_LATEST",
	}
	VersionType_value = map[string]int32{
		"VERSION_TYPE_UNSPECIFIED": 0,
		"VERSION_TYPE_PINNED":      1,
		"VERSION_TYPE_LATEST":      2,
	}
)

func (x VersionType) Enum() *VersionType {
	p := new(VersionType)
	*p = x
	return p
}

func (x VersionType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (VersionType) Descriptor() protoreflect.EnumDescriptor {
	return file_plantr_controller_v1_struct_proto_enumTypes[0].Descriptor()
}

func (VersionType) Type() protoreflect.EnumType {
	return &file_plantr_controller_v1_struct_proto_enumTypes[0]
}

func (x VersionType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use VersionType.Descriptor instead.
func (VersionType) EnumDescriptor() ([]byte, []int) {
	return file_plantr_controller_v1_struct_proto_rawDescGZIP(), []int{0}
}

type ConfigFile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content     string `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Destination string `protobuf:"bytes,2,opt,name=destination,proto3" json:"destination,omitempty"`
}

func (x *ConfigFile) Reset() {
	*x = ConfigFile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plantr_controller_v1_struct_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigFile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigFile) ProtoMessage() {}

func (x *ConfigFile) ProtoReflect() protoreflect.Message {
	mi := &file_plantr_controller_v1_struct_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigFile.ProtoReflect.Descriptor instead.
func (*ConfigFile) Descriptor() ([]byte, []int) {
	return file_plantr_controller_v1_struct_proto_rawDescGZIP(), []int{0}
}

func (x *ConfigFile) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *ConfigFile) GetDestination() string {
	if x != nil {
		return x.Destination
	}
	return ""
}

type Seed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Element:
	//
	//	*Seed_ConfigFile
	Element isSeed_Element `protobuf_oneof:"element"`
}

func (x *Seed) Reset() {
	*x = Seed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plantr_controller_v1_struct_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Seed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Seed) ProtoMessage() {}

func (x *Seed) ProtoReflect() protoreflect.Message {
	mi := &file_plantr_controller_v1_struct_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Seed.ProtoReflect.Descriptor instead.
func (*Seed) Descriptor() ([]byte, []int) {
	return file_plantr_controller_v1_struct_proto_rawDescGZIP(), []int{1}
}

func (m *Seed) GetElement() isSeed_Element {
	if m != nil {
		return m.Element
	}
	return nil
}

func (x *Seed) GetConfigFile() *ConfigFile {
	if x, ok := x.GetElement().(*Seed_ConfigFile); ok {
		return x.ConfigFile
	}
	return nil
}

type isSeed_Element interface {
	isSeed_Element()
}

type Seed_ConfigFile struct {
	ConfigFile *ConfigFile `protobuf:"bytes,1,opt,name=config_file,json=configFile,proto3,oneof"`
}

func (*Seed_ConfigFile) isSeed_Element() {}

var File_plantr_controller_v1_struct_proto protoreflect.FileDescriptor

var file_plantr_controller_v1_struct_proto_rawDesc = []byte{
	0x0a, 0x21, 0x70, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x14, 0x70, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x22, 0x48, 0x0a, 0x0a, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x22, 0x56, 0x0a, 0x04, 0x53, 0x65, 0x65, 0x64, 0x12, 0x43, 0x0a, 0x0b, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x70, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x46, 0x69,
	0x6c, 0x65, 0x48, 0x00, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x46, 0x69, 0x6c, 0x65,
	0x42, 0x09, 0x0a, 0x07, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x2a, 0x5d, 0x0a, 0x0b, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x18, 0x56, 0x45,
	0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45,
	0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x17, 0x0a, 0x13, 0x56, 0x45, 0x52, 0x53,
	0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50, 0x49, 0x4e, 0x4e, 0x45, 0x44, 0x10,
	0x01, 0x12, 0x17, 0x0a, 0x13, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50,
	0x45, 0x5f, 0x4c, 0x41, 0x54, 0x45, 0x53, 0x54, 0x10, 0x02, 0x42, 0xe0, 0x01, 0x0a, 0x18, 0x63,
	0x6f, 0x6d, 0x2e, 0x70, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6e, 0x69, 0x63, 0x6a, 0x6f, 0x68, 0x6e, 0x73, 0x6f, 0x6e, 0x31, 0x34, 0x35,
	0x2f, 0x70, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x6c, 0x61, 0x6e,
	0x74, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x76, 0x31,
	0x3b, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x76, 0x31, 0xa2, 0x02, 0x03,
	0x50, 0x43, 0x58, 0xaa, 0x02, 0x14, 0x50, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x2e, 0x43, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x14, 0x50, 0x6c, 0x61,
	0x6e, 0x74, 0x72, 0x5c, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x5c, 0x56,
	0x31, 0xe2, 0x02, 0x20, 0x50, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x5c, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x16, 0x50, 0x6c, 0x61, 0x6e, 0x74, 0x72, 0x3a, 0x3a, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_plantr_controller_v1_struct_proto_rawDescOnce sync.Once
	file_plantr_controller_v1_struct_proto_rawDescData = file_plantr_controller_v1_struct_proto_rawDesc
)

func file_plantr_controller_v1_struct_proto_rawDescGZIP() []byte {
	file_plantr_controller_v1_struct_proto_rawDescOnce.Do(func() {
		file_plantr_controller_v1_struct_proto_rawDescData = protoimpl.X.CompressGZIP(file_plantr_controller_v1_struct_proto_rawDescData)
	})
	return file_plantr_controller_v1_struct_proto_rawDescData
}

var file_plantr_controller_v1_struct_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_plantr_controller_v1_struct_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_plantr_controller_v1_struct_proto_goTypes = []any{
	(VersionType)(0),   // 0: plantr.controller.v1.VersionType
	(*ConfigFile)(nil), // 1: plantr.controller.v1.ConfigFile
	(*Seed)(nil),       // 2: plantr.controller.v1.Seed
}
var file_plantr_controller_v1_struct_proto_depIdxs = []int32{
	1, // 0: plantr.controller.v1.Seed.config_file:type_name -> plantr.controller.v1.ConfigFile
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_plantr_controller_v1_struct_proto_init() }
func file_plantr_controller_v1_struct_proto_init() {
	if File_plantr_controller_v1_struct_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_plantr_controller_v1_struct_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*ConfigFile); i {
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
		file_plantr_controller_v1_struct_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Seed); i {
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
	file_plantr_controller_v1_struct_proto_msgTypes[1].OneofWrappers = []any{
		(*Seed_ConfigFile)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_plantr_controller_v1_struct_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_plantr_controller_v1_struct_proto_goTypes,
		DependencyIndexes: file_plantr_controller_v1_struct_proto_depIdxs,
		EnumInfos:         file_plantr_controller_v1_struct_proto_enumTypes,
		MessageInfos:      file_plantr_controller_v1_struct_proto_msgTypes,
	}.Build()
	File_plantr_controller_v1_struct_proto = out.File
	file_plantr_controller_v1_struct_proto_rawDesc = nil
	file_plantr_controller_v1_struct_proto_goTypes = nil
	file_plantr_controller_v1_struct_proto_depIdxs = nil
}
