// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: store/user.proto

package store

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// MFAConfig is the MFA configuration for a user.
type MFAConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The otp_secret is the secret key used to validate the OTP code.
	OtpSecret string `protobuf:"bytes,1,opt,name=otp_secret,json=otpSecret,proto3" json:"otp_secret,omitempty"`
	// The temp_otp_secret is the temporary secret key used to validate the OTP code and will replace the otp_secret in two phase commits.
	TempOtpSecret string `protobuf:"bytes,2,opt,name=temp_otp_secret,json=tempOtpSecret,proto3" json:"temp_otp_secret,omitempty"`
	// The recovery_codes are the codes that can be used to recover the account.
	RecoveryCodes []string `protobuf:"bytes,3,rep,name=recovery_codes,json=recoveryCodes,proto3" json:"recovery_codes,omitempty"`
	// The temp_recovery_codes are the temporary codes that will replace the recovery_codes in two phase commits.
	TempRecoveryCodes []string `protobuf:"bytes,4,rep,name=temp_recovery_codes,json=tempRecoveryCodes,proto3" json:"temp_recovery_codes,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *MFAConfig) Reset() {
	*x = MFAConfig{}
	mi := &file_store_user_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MFAConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MFAConfig) ProtoMessage() {}

func (x *MFAConfig) ProtoReflect() protoreflect.Message {
	mi := &file_store_user_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MFAConfig.ProtoReflect.Descriptor instead.
func (*MFAConfig) Descriptor() ([]byte, []int) {
	return file_store_user_proto_rawDescGZIP(), []int{0}
}

func (x *MFAConfig) GetOtpSecret() string {
	if x != nil {
		return x.OtpSecret
	}
	return ""
}

func (x *MFAConfig) GetTempOtpSecret() string {
	if x != nil {
		return x.TempOtpSecret
	}
	return ""
}

func (x *MFAConfig) GetRecoveryCodes() []string {
	if x != nil {
		return x.RecoveryCodes
	}
	return nil
}

func (x *MFAConfig) GetTempRecoveryCodes() []string {
	if x != nil {
		return x.TempRecoveryCodes
	}
	return nil
}

type UserProfile struct {
	state                  protoimpl.MessageState `protogen:"open.v1"`
	LastLoginTime          *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=last_login_time,json=lastLoginTime,proto3" json:"last_login_time,omitempty"`
	LastChangePasswordTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=last_change_password_time,json=lastChangePasswordTime,proto3" json:"last_change_password_time,omitempty"`
	// source means where the user comes from. For now we support Entra ID SCIM sync, so the source could be Entra ID.
	Source        string `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserProfile) Reset() {
	*x = UserProfile{}
	mi := &file_store_user_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserProfile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserProfile) ProtoMessage() {}

func (x *UserProfile) ProtoReflect() protoreflect.Message {
	mi := &file_store_user_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserProfile.ProtoReflect.Descriptor instead.
func (*UserProfile) Descriptor() ([]byte, []int) {
	return file_store_user_proto_rawDescGZIP(), []int{1}
}

func (x *UserProfile) GetLastLoginTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LastLoginTime
	}
	return nil
}

func (x *UserProfile) GetLastChangePasswordTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LastChangePasswordTime
	}
	return nil
}

func (x *UserProfile) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

var File_store_user_proto protoreflect.FileDescriptor

const file_store_user_proto_rawDesc = "" +
	"\n" +
	"\x10store/user.proto\x12\x0ebytebase.store\x1a\x1fgoogle/protobuf/timestamp.proto\"\xa9\x01\n" +
	"\tMFAConfig\x12\x1d\n" +
	"\n" +
	"otp_secret\x18\x01 \x01(\tR\totpSecret\x12&\n" +
	"\x0ftemp_otp_secret\x18\x02 \x01(\tR\rtempOtpSecret\x12%\n" +
	"\x0erecovery_codes\x18\x03 \x03(\tR\rrecoveryCodes\x12.\n" +
	"\x13temp_recovery_codes\x18\x04 \x03(\tR\x11tempRecoveryCodes\"\xc0\x01\n" +
	"\vUserProfile\x12B\n" +
	"\x0flast_login_time\x18\x01 \x01(\v2\x1a.google.protobuf.TimestampR\rlastLoginTime\x12U\n" +
	"\x19last_change_password_time\x18\x02 \x01(\v2\x1a.google.protobuf.TimestampR\x16lastChangePasswordTime\x12\x16\n" +
	"\x06source\x18\x03 \x01(\tR\x06sourceB\x14Z\x12generated-go/storeb\x06proto3"

var (
	file_store_user_proto_rawDescOnce sync.Once
	file_store_user_proto_rawDescData []byte
)

func file_store_user_proto_rawDescGZIP() []byte {
	file_store_user_proto_rawDescOnce.Do(func() {
		file_store_user_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_store_user_proto_rawDesc), len(file_store_user_proto_rawDesc)))
	})
	return file_store_user_proto_rawDescData
}

var file_store_user_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_store_user_proto_goTypes = []any{
	(*MFAConfig)(nil),             // 0: bytebase.store.MFAConfig
	(*UserProfile)(nil),           // 1: bytebase.store.UserProfile
	(*timestamppb.Timestamp)(nil), // 2: google.protobuf.Timestamp
}
var file_store_user_proto_depIdxs = []int32{
	2, // 0: bytebase.store.UserProfile.last_login_time:type_name -> google.protobuf.Timestamp
	2, // 1: bytebase.store.UserProfile.last_change_password_time:type_name -> google.protobuf.Timestamp
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_store_user_proto_init() }
func file_store_user_proto_init() {
	if File_store_user_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_store_user_proto_rawDesc), len(file_store_user_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_store_user_proto_goTypes,
		DependencyIndexes: file_store_user_proto_depIdxs,
		MessageInfos:      file_store_user_proto_msgTypes,
	}.Build()
	File_store_user_proto = out.File
	file_store_user_proto_goTypes = nil
	file_store_user_proto_depIdxs = nil
}
