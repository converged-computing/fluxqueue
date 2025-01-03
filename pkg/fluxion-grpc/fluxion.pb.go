// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.3
// source: pkg/fluxion-grpc/fluxion.proto

package fluxion_grpc

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

type PodSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Container string   `protobuf:"bytes,2,opt,name=container,proto3" json:"container,omitempty"`
	Cpu       int32    `protobuf:"varint,3,opt,name=cpu,proto3" json:"cpu,omitempty"`
	Memory    int64    `protobuf:"varint,4,opt,name=memory,proto3" json:"memory,omitempty"`
	Gpu       int64    `protobuf:"varint,5,opt,name=gpu,proto3" json:"gpu,omitempty"`
	Storage   int64    `protobuf:"varint,6,opt,name=storage,proto3" json:"storage,omitempty"`
	Labels    []string `protobuf:"bytes,7,rep,name=labels,proto3" json:"labels,omitempty"`
}

func (x *PodSpec) Reset() {
	*x = PodSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PodSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PodSpec) ProtoMessage() {}

func (x *PodSpec) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PodSpec.ProtoReflect.Descriptor instead.
func (*PodSpec) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{0}
}

func (x *PodSpec) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *PodSpec) GetContainer() string {
	if x != nil {
		return x.Container
	}
	return ""
}

func (x *PodSpec) GetCpu() int32 {
	if x != nil {
		return x.Cpu
	}
	return 0
}

func (x *PodSpec) GetMemory() int64 {
	if x != nil {
		return x.Memory
	}
	return 0
}

func (x *PodSpec) GetGpu() int64 {
	if x != nil {
		return x.Gpu
	}
	return 0
}

func (x *PodSpec) GetStorage() int64 {
	if x != nil {
		return x.Storage
	}
	return 0
}

func (x *PodSpec) GetLabels() []string {
	if x != nil {
		return x.Labels
	}
	return nil
}

// The Match request message (allocate, allocate_orelse_reserve)
// TODO: this currently takes a podspec, and we multiply by a count
// we should ideally support having a list of different pods
type MatchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Podspec *PodSpec `protobuf:"bytes,1,opt,name=podspec,proto3" json:"podspec,omitempty"`
	Count   int32    `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Reserve bool     `protobuf:"varint,4,opt,name=reserve,proto3" json:"reserve,omitempty"`
	JobName string   `protobuf:"bytes,5,opt,name=jobName,proto3" json:"jobName,omitempty"`
}

func (x *MatchRequest) Reset() {
	*x = MatchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MatchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MatchRequest) ProtoMessage() {}

func (x *MatchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MatchRequest.ProtoReflect.Descriptor instead.
func (*MatchRequest) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{1}
}

func (x *MatchRequest) GetPodspec() *PodSpec {
	if x != nil {
		return x.Podspec
	}
	return nil
}

func (x *MatchRequest) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *MatchRequest) GetReserve() bool {
	if x != nil {
		return x.Reserve
	}
	return false
}

func (x *MatchRequest) GetJobName() string {
	if x != nil {
		return x.JobName
	}
	return ""
}

// The Nodes/Cluster Update Status
type NodeAlloc struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeID string `protobuf:"bytes,1,opt,name=nodeID,proto3" json:"nodeID,omitempty"`
	Tasks  int32  `protobuf:"varint,2,opt,name=tasks,proto3" json:"tasks,omitempty"`
}

func (x *NodeAlloc) Reset() {
	*x = NodeAlloc{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeAlloc) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeAlloc) ProtoMessage() {}

func (x *NodeAlloc) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeAlloc.ProtoReflect.Descriptor instead.
func (*NodeAlloc) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{2}
}

func (x *NodeAlloc) GetNodeID() string {
	if x != nil {
		return x.NodeID
	}
	return ""
}

func (x *NodeAlloc) GetTasks() int32 {
	if x != nil {
		return x.Tasks
	}
	return 0
}

// The Match response message
type MatchResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FluxID     uint64       `protobuf:"varint,1,opt,name=fluxID,proto3" json:"fluxID,omitempty"`
	Nodelist   []*NodeAlloc `protobuf:"bytes,2,rep,name=nodelist,proto3" json:"nodelist,omitempty"`
	Reserved   bool         `protobuf:"varint,3,opt,name=reserved,proto3" json:"reserved,omitempty"`
	ReservedAt int64        `protobuf:"varint,4,opt,name=reserved_at,json=reservedAt,proto3" json:"reserved_at,omitempty"`
	// Only needed if we want stats or similar
	// float overhead = 5;
	// boolean to indicate allocated or not
	Allocated bool `protobuf:"varint,5,opt,name=allocated,proto3" json:"allocated,omitempty"`
}

func (x *MatchResponse) Reset() {
	*x = MatchResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MatchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MatchResponse) ProtoMessage() {}

func (x *MatchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MatchResponse.ProtoReflect.Descriptor instead.
func (*MatchResponse) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{3}
}

func (x *MatchResponse) GetFluxID() uint64 {
	if x != nil {
		return x.FluxID
	}
	return 0
}

func (x *MatchResponse) GetNodelist() []*NodeAlloc {
	if x != nil {
		return x.Nodelist
	}
	return nil
}

func (x *MatchResponse) GetReserved() bool {
	if x != nil {
		return x.Reserved
	}
	return false
}

func (x *MatchResponse) GetReservedAt() int64 {
	if x != nil {
		return x.ReservedAt
	}
	return 0
}

func (x *MatchResponse) GetAllocated() bool {
	if x != nil {
		return x.Allocated
	}
	return false
}

type CancelRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FluxID uint64 `protobuf:"varint,1,opt,name=fluxID,proto3" json:"fluxID,omitempty"`
	// It's ok if it doesn't exist (don't issue an error)
	NoExistOK bool `protobuf:"varint,2,opt,name=NoExistOK,proto3" json:"NoExistOK,omitempty"`
}

func (x *CancelRequest) Reset() {
	*x = CancelRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CancelRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelRequest) ProtoMessage() {}

func (x *CancelRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelRequest.ProtoReflect.Descriptor instead.
func (*CancelRequest) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{4}
}

func (x *CancelRequest) GetFluxID() uint64 {
	if x != nil {
		return x.FluxID
	}
	return 0
}

func (x *CancelRequest) GetNoExistOK() bool {
	if x != nil {
		return x.NoExistOK
	}
	return false
}

// The Match response message
type CancelResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FluxID uint64 `protobuf:"varint,1,opt,name=fluxID,proto3" json:"fluxID,omitempty"`
	Error  int32  `protobuf:"varint,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *CancelResponse) Reset() {
	*x = CancelResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CancelResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelResponse) ProtoMessage() {}

func (x *CancelResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelResponse.ProtoReflect.Descriptor instead.
func (*CancelResponse) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{5}
}

func (x *CancelResponse) GetFluxID() uint64 {
	if x != nil {
		return x.FluxID
	}
	return 0
}

func (x *CancelResponse) GetError() int32 {
	if x != nil {
		return x.Error
	}
	return 0
}

// The Nodes/Cluster Update Status
type NodeStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CpuAvail     int32  `protobuf:"varint,1,opt,name=cpuAvail,proto3" json:"cpuAvail,omitempty"`
	GpuAvail     int32  `protobuf:"varint,2,opt,name=gpuAvail,proto3" json:"gpuAvail,omitempty"`
	StorageAvail int64  `protobuf:"varint,3,opt,name=storageAvail,proto3" json:"storageAvail,omitempty"`
	MemoryAvail  int64  `protobuf:"varint,4,opt,name=memoryAvail,proto3" json:"memoryAvail,omitempty"`
	AllowedPods  int64  `protobuf:"varint,5,opt,name=allowedPods,proto3" json:"allowedPods,omitempty"`
	NodeIP       string `protobuf:"bytes,6,opt,name=nodeIP,proto3" json:"nodeIP,omitempty"`
	Replication  int32  `protobuf:"varint,7,opt,name=replication,proto3" json:"replication,omitempty"`
}

func (x *NodeStatus) Reset() {
	*x = NodeStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeStatus) ProtoMessage() {}

func (x *NodeStatus) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeStatus.ProtoReflect.Descriptor instead.
func (*NodeStatus) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{6}
}

func (x *NodeStatus) GetCpuAvail() int32 {
	if x != nil {
		return x.CpuAvail
	}
	return 0
}

func (x *NodeStatus) GetGpuAvail() int32 {
	if x != nil {
		return x.GpuAvail
	}
	return 0
}

func (x *NodeStatus) GetStorageAvail() int64 {
	if x != nil {
		return x.StorageAvail
	}
	return 0
}

func (x *NodeStatus) GetMemoryAvail() int64 {
	if x != nil {
		return x.MemoryAvail
	}
	return 0
}

func (x *NodeStatus) GetAllowedPods() int64 {
	if x != nil {
		return x.AllowedPods
	}
	return 0
}

func (x *NodeStatus) GetNodeIP() string {
	if x != nil {
		return x.NodeIP
	}
	return ""
}

func (x *NodeStatus) GetReplication() int32 {
	if x != nil {
		return x.Replication
	}
	return 0
}

// The JGF response message
type JGFRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Jgf string `protobuf:"bytes,1,opt,name=jgf,proto3" json:"jgf,omitempty"`
}

func (x *JGFRequest) Reset() {
	*x = JGFRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JGFRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JGFRequest) ProtoMessage() {}

func (x *JGFRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JGFRequest.ProtoReflect.Descriptor instead.
func (*JGFRequest) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{7}
}

func (x *JGFRequest) GetJgf() string {
	if x != nil {
		return x.Jgf
	}
	return ""
}

// The JGF response message
type JGFResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Jgf string `protobuf:"bytes,1,opt,name=jgf,proto3" json:"jgf,omitempty"`
}

func (x *JGFResponse) Reset() {
	*x = JGFResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JGFResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JGFResponse) ProtoMessage() {}

func (x *JGFResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_fluxion_grpc_fluxion_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JGFResponse.ProtoReflect.Descriptor instead.
func (*JGFResponse) Descriptor() ([]byte, []int) {
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP(), []int{8}
}

func (x *JGFResponse) GetJgf() string {
	if x != nil {
		return x.Jgf
	}
	return ""
}

var File_pkg_fluxion_grpc_fluxion_proto protoreflect.FileDescriptor

var file_pkg_fluxion_grpc_fluxion_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x70, 0x6b, 0x67, 0x2f, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2d, 0x67, 0x72,
	0x70, 0x63, 0x2f, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x22, 0xa5, 0x01, 0x0a, 0x07, 0x50, 0x6f,
	0x64, 0x53, 0x70, 0x65, 0x63, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x70, 0x75, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x03, 0x63, 0x70, 0x75, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x67, 0x70, 0x75, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x67, 0x70, 0x75, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x61, 0x62,
	0x65, 0x6c, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c,
	0x73, 0x22, 0x84, 0x01, 0x0a, 0x0c, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x2a, 0x0a, 0x07, 0x70, 0x6f, 0x64, 0x73, 0x70, 0x65, 0x63, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x6f,
	0x64, 0x53, 0x70, 0x65, 0x63, 0x52, 0x07, 0x70, 0x6f, 0x64, 0x73, 0x70, 0x65, 0x63, 0x12, 0x14,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x6a, 0x6f, 0x62, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x6a, 0x6f, 0x62, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x39, 0x0a, 0x09, 0x4e, 0x6f, 0x64, 0x65,
	0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x12, 0x14, 0x0a,
	0x05, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x74, 0x61,
	0x73, 0x6b, 0x73, 0x22, 0xb2, 0x01, 0x0a, 0x0d, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6c, 0x75, 0x78, 0x49, 0x44, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x66, 0x6c, 0x75, 0x78, 0x49, 0x44, 0x12, 0x2e, 0x0a,
	0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x41, 0x6c,
	0x6c, 0x6f, 0x63, 0x52, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x08, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a,
	0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x6c,
	0x6c, 0x6f, 0x63, 0x61, 0x74, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x61,
	0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x65, 0x64, 0x22, 0x45, 0x0a, 0x0d, 0x43, 0x61, 0x6e, 0x63,
	0x65, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6c, 0x75,
	0x78, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x66, 0x6c, 0x75, 0x78, 0x49,
	0x44, 0x12, 0x1c, 0x0a, 0x09, 0x4e, 0x6f, 0x45, 0x78, 0x69, 0x73, 0x74, 0x4f, 0x4b, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x4e, 0x6f, 0x45, 0x78, 0x69, 0x73, 0x74, 0x4f, 0x4b, 0x22,
	0x3e, 0x0a, 0x0e, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6c, 0x75, 0x78, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x06, 0x66, 0x6c, 0x75, 0x78, 0x49, 0x44, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22,
	0xe6, 0x01, 0x0a, 0x0a, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a,
	0x0a, 0x08, 0x63, 0x70, 0x75, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x08, 0x63, 0x70, 0x75, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x67, 0x70,
	0x75, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x67, 0x70,
	0x75, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x12, 0x20, 0x0a, 0x0b, 0x6d, 0x65,
	0x6d, 0x6f, 0x72, 0x79, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0b, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x12, 0x20, 0x0a, 0x0b,
	0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x50, 0x6f, 0x64, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0b, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x50, 0x6f, 0x64, 0x73, 0x12, 0x16,
	0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x50, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x6e, 0x6f, 0x64, 0x65, 0x49, 0x50, 0x12, 0x20, 0x0a, 0x0b, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x72, 0x65, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x1e, 0x0a, 0x0a, 0x4a, 0x47, 0x46, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6a, 0x67, 0x66, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6a, 0x67, 0x66, 0x22, 0x1f, 0x0a, 0x0b, 0x4a, 0x47, 0x46, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6a, 0x67, 0x66, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6a, 0x67, 0x66, 0x32, 0x87, 0x01, 0x0a, 0x0e, 0x46, 0x6c,
	0x75, 0x78, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x38, 0x0a, 0x05,
	0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x15, 0x2e, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2e,
	0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x66,
	0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3b, 0x0a, 0x06, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c,
	0x12, 0x16, 0x2e, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x61, 0x6e, 0x63, 0x65,
	0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x66, 0x6c, 0x75, 0x78, 0x69,
	0x6f, 0x6e, 0x2e, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x3b, 0x5a, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x67, 0x65, 0x64, 0x2d, 0x63, 0x6f, 0x6d, 0x70,
	0x75, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x66, 0x6c, 0x75, 0x78, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x66, 0x6c, 0x75, 0x78, 0x69, 0x6f, 0x6e, 0x2d, 0x67, 0x72, 0x70, 0x63,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_fluxion_grpc_fluxion_proto_rawDescOnce sync.Once
	file_pkg_fluxion_grpc_fluxion_proto_rawDescData = file_pkg_fluxion_grpc_fluxion_proto_rawDesc
)

func file_pkg_fluxion_grpc_fluxion_proto_rawDescGZIP() []byte {
	file_pkg_fluxion_grpc_fluxion_proto_rawDescOnce.Do(func() {
		file_pkg_fluxion_grpc_fluxion_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_fluxion_grpc_fluxion_proto_rawDescData)
	})
	return file_pkg_fluxion_grpc_fluxion_proto_rawDescData
}

var file_pkg_fluxion_grpc_fluxion_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_pkg_fluxion_grpc_fluxion_proto_goTypes = []interface{}{
	(*PodSpec)(nil),        // 0: fluxion.PodSpec
	(*MatchRequest)(nil),   // 1: fluxion.MatchRequest
	(*NodeAlloc)(nil),      // 2: fluxion.NodeAlloc
	(*MatchResponse)(nil),  // 3: fluxion.MatchResponse
	(*CancelRequest)(nil),  // 4: fluxion.CancelRequest
	(*CancelResponse)(nil), // 5: fluxion.CancelResponse
	(*NodeStatus)(nil),     // 6: fluxion.NodeStatus
	(*JGFRequest)(nil),     // 7: fluxion.JGFRequest
	(*JGFResponse)(nil),    // 8: fluxion.JGFResponse
}
var file_pkg_fluxion_grpc_fluxion_proto_depIdxs = []int32{
	0, // 0: fluxion.MatchRequest.podspec:type_name -> fluxion.PodSpec
	2, // 1: fluxion.MatchResponse.nodelist:type_name -> fluxion.NodeAlloc
	1, // 2: fluxion.FluxionService.Match:input_type -> fluxion.MatchRequest
	4, // 3: fluxion.FluxionService.Cancel:input_type -> fluxion.CancelRequest
	3, // 4: fluxion.FluxionService.Match:output_type -> fluxion.MatchResponse
	5, // 5: fluxion.FluxionService.Cancel:output_type -> fluxion.CancelResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pkg_fluxion_grpc_fluxion_proto_init() }
func file_pkg_fluxion_grpc_fluxion_proto_init() {
	if File_pkg_fluxion_grpc_fluxion_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PodSpec); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MatchRequest); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeAlloc); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MatchResponse); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CancelRequest); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CancelResponse); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeStatus); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JGFRequest); i {
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
		file_pkg_fluxion_grpc_fluxion_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JGFResponse); i {
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
			RawDescriptor: file_pkg_fluxion_grpc_fluxion_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_fluxion_grpc_fluxion_proto_goTypes,
		DependencyIndexes: file_pkg_fluxion_grpc_fluxion_proto_depIdxs,
		MessageInfos:      file_pkg_fluxion_grpc_fluxion_proto_msgTypes,
	}.Build()
	File_pkg_fluxion_grpc_fluxion_proto = out.File
	file_pkg_fluxion_grpc_fluxion_proto_rawDesc = nil
	file_pkg_fluxion_grpc_fluxion_proto_goTypes = nil
	file_pkg_fluxion_grpc_fluxion_proto_depIdxs = nil
}
