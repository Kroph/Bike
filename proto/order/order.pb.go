// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: proto/order/order.proto

package order

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

type OrderStatus int32

const (
	OrderStatus_PENDING   OrderStatus = 0
	OrderStatus_PAID      OrderStatus = 1
	OrderStatus_SHIPPED   OrderStatus = 2
	OrderStatus_DELIVERED OrderStatus = 3
	OrderStatus_CANCELLED OrderStatus = 4
)

// Enum value maps for OrderStatus.
var (
	OrderStatus_name = map[int32]string{
		0: "PENDING",
		1: "PAID",
		2: "SHIPPED",
		3: "DELIVERED",
		4: "CANCELLED",
	}
	OrderStatus_value = map[string]int32{
		"PENDING":   0,
		"PAID":      1,
		"SHIPPED":   2,
		"DELIVERED": 3,
		"CANCELLED": 4,
	}
)

func (x OrderStatus) Enum() *OrderStatus {
	p := new(OrderStatus)
	*p = x
	return p
}

func (x OrderStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OrderStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_order_order_proto_enumTypes[0].Descriptor()
}

func (OrderStatus) Type() protoreflect.EnumType {
	return &file_proto_order_order_proto_enumTypes[0]
}

func (x OrderStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OrderStatus.Descriptor instead.
func (OrderStatus) EnumDescriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{0}
}

type CreateOrderRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Items         []*OrderItemRequest    `protobuf:"bytes,2,rep,name=items,proto3" json:"items,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateOrderRequest) Reset() {
	*x = CreateOrderRequest{}
	mi := &file_proto_order_order_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateOrderRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateOrderRequest) ProtoMessage() {}

func (x *CreateOrderRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateOrderRequest.ProtoReflect.Descriptor instead.
func (*CreateOrderRequest) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{0}
}

func (x *CreateOrderRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *CreateOrderRequest) GetItems() []*OrderItemRequest {
	if x != nil {
		return x.Items
	}
	return nil
}

type OrderResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Status        OrderStatus            `protobuf:"varint,3,opt,name=status,proto3,enum=order.OrderStatus" json:"status,omitempty"`
	Total         float64                `protobuf:"fixed64,4,opt,name=total,proto3" json:"total,omitempty"`
	Items         []*OrderItemResponse   `protobuf:"bytes,5,rep,name=items,proto3" json:"items,omitempty"`
	CreatedAt     *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt     *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OrderResponse) Reset() {
	*x = OrderResponse{}
	mi := &file_proto_order_order_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OrderResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderResponse) ProtoMessage() {}

func (x *OrderResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderResponse.ProtoReflect.Descriptor instead.
func (*OrderResponse) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{1}
}

func (x *OrderResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *OrderResponse) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *OrderResponse) GetStatus() OrderStatus {
	if x != nil {
		return x.Status
	}
	return OrderStatus_PENDING
}

func (x *OrderResponse) GetTotal() float64 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *OrderResponse) GetItems() []*OrderItemResponse {
	if x != nil {
		return x.Items
	}
	return nil
}

func (x *OrderResponse) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *OrderResponse) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

type OrderIDRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OrderIDRequest) Reset() {
	*x = OrderIDRequest{}
	mi := &file_proto_order_order_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OrderIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderIDRequest) ProtoMessage() {}

func (x *OrderIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderIDRequest.ProtoReflect.Descriptor instead.
func (*OrderIDRequest) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{2}
}

func (x *OrderIDRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type UserIDRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserIDRequest) Reset() {
	*x = UserIDRequest{}
	mi := &file_proto_order_order_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserIDRequest) ProtoMessage() {}

func (x *UserIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserIDRequest.ProtoReflect.Descriptor instead.
func (*UserIDRequest) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{3}
}

func (x *UserIDRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type UpdateOrderStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Status        OrderStatus            `protobuf:"varint,2,opt,name=status,proto3,enum=order.OrderStatus" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateOrderStatusRequest) Reset() {
	*x = UpdateOrderStatusRequest{}
	mi := &file_proto_order_order_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateOrderStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateOrderStatusRequest) ProtoMessage() {}

func (x *UpdateOrderStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateOrderStatusRequest.ProtoReflect.Descriptor instead.
func (*UpdateOrderStatusRequest) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{4}
}

func (x *UpdateOrderStatusRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *UpdateOrderStatusRequest) GetStatus() OrderStatus {
	if x != nil {
		return x.Status
	}
	return OrderStatus_PENDING
}

type OrderFilter struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserId        string                 `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Status        OrderStatus            `protobuf:"varint,2,opt,name=status,proto3,enum=order.OrderStatus" json:"status,omitempty"`
	FromDate      *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=from_date,json=fromDate,proto3" json:"from_date,omitempty"`
	ToDate        *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=to_date,json=toDate,proto3" json:"to_date,omitempty"`
	Page          int32                  `protobuf:"varint,5,opt,name=page,proto3" json:"page,omitempty"`
	PageSize      int32                  `protobuf:"varint,6,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OrderFilter) Reset() {
	*x = OrderFilter{}
	mi := &file_proto_order_order_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OrderFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderFilter) ProtoMessage() {}

func (x *OrderFilter) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderFilter.ProtoReflect.Descriptor instead.
func (*OrderFilter) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{5}
}

func (x *OrderFilter) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *OrderFilter) GetStatus() OrderStatus {
	if x != nil {
		return x.Status
	}
	return OrderStatus_PENDING
}

func (x *OrderFilter) GetFromDate() *timestamppb.Timestamp {
	if x != nil {
		return x.FromDate
	}
	return nil
}

func (x *OrderFilter) GetToDate() *timestamppb.Timestamp {
	if x != nil {
		return x.ToDate
	}
	return nil
}

func (x *OrderFilter) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *OrderFilter) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type ListOrdersRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Filter        *OrderFilter           `protobuf:"bytes,1,opt,name=filter,proto3" json:"filter,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListOrdersRequest) Reset() {
	*x = ListOrdersRequest{}
	mi := &file_proto_order_order_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListOrdersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListOrdersRequest) ProtoMessage() {}

func (x *ListOrdersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListOrdersRequest.ProtoReflect.Descriptor instead.
func (*ListOrdersRequest) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{6}
}

func (x *ListOrdersRequest) GetFilter() *OrderFilter {
	if x != nil {
		return x.Filter
	}
	return nil
}

type ListOrdersResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Orders        []*OrderResponse       `protobuf:"bytes,1,rep,name=orders,proto3" json:"orders,omitempty"`
	Total         int32                  `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
	Page          int32                  `protobuf:"varint,3,opt,name=page,proto3" json:"page,omitempty"`
	PageSize      int32                  `protobuf:"varint,4,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListOrdersResponse) Reset() {
	*x = ListOrdersResponse{}
	mi := &file_proto_order_order_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListOrdersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListOrdersResponse) ProtoMessage() {}

func (x *ListOrdersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListOrdersResponse.ProtoReflect.Descriptor instead.
func (*ListOrdersResponse) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{7}
}

func (x *ListOrdersResponse) GetOrders() []*OrderResponse {
	if x != nil {
		return x.Orders
	}
	return nil
}

func (x *ListOrdersResponse) GetTotal() int32 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *ListOrdersResponse) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *ListOrdersResponse) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type OrderItemRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ProductId     string                 `protobuf:"bytes,1,opt,name=product_id,json=productId,proto3" json:"product_id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Price         float64                `protobuf:"fixed64,3,opt,name=price,proto3" json:"price,omitempty"`
	Quantity      int32                  `protobuf:"varint,4,opt,name=quantity,proto3" json:"quantity,omitempty"`
	FrameSize     string                 `protobuf:"bytes,5,opt,name=frame_size,json=frameSize,proto3" json:"frame_size,omitempty"`
	WheelSize     string                 `protobuf:"bytes,6,opt,name=wheel_size,json=wheelSize,proto3" json:"wheel_size,omitempty"`
	Color         string                 `protobuf:"bytes,7,opt,name=color,proto3" json:"color,omitempty"`
	BikeType      string                 `protobuf:"bytes,8,opt,name=bike_type,json=bikeType,proto3" json:"bike_type,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OrderItemRequest) Reset() {
	*x = OrderItemRequest{}
	mi := &file_proto_order_order_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OrderItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderItemRequest) ProtoMessage() {}

func (x *OrderItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderItemRequest.ProtoReflect.Descriptor instead.
func (*OrderItemRequest) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{8}
}

func (x *OrderItemRequest) GetProductId() string {
	if x != nil {
		return x.ProductId
	}
	return ""
}

func (x *OrderItemRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *OrderItemRequest) GetPrice() float64 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *OrderItemRequest) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

func (x *OrderItemRequest) GetFrameSize() string {
	if x != nil {
		return x.FrameSize
	}
	return ""
}

func (x *OrderItemRequest) GetWheelSize() string {
	if x != nil {
		return x.WheelSize
	}
	return ""
}

func (x *OrderItemRequest) GetColor() string {
	if x != nil {
		return x.Color
	}
	return ""
}

func (x *OrderItemRequest) GetBikeType() string {
	if x != nil {
		return x.BikeType
	}
	return ""
}

type OrderItemResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	OrderId       string                 `protobuf:"bytes,2,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	ProductId     string                 `protobuf:"bytes,3,opt,name=product_id,json=productId,proto3" json:"product_id,omitempty"`
	Name          string                 `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Price         float64                `protobuf:"fixed64,5,opt,name=price,proto3" json:"price,omitempty"`
	Quantity      int32                  `protobuf:"varint,6,opt,name=quantity,proto3" json:"quantity,omitempty"`
	FrameSize     string                 `protobuf:"bytes,7,opt,name=frame_size,json=frameSize,proto3" json:"frame_size,omitempty"`
	WheelSize     string                 `protobuf:"bytes,8,opt,name=wheel_size,json=wheelSize,proto3" json:"wheel_size,omitempty"`
	Color         string                 `protobuf:"bytes,9,opt,name=color,proto3" json:"color,omitempty"`
	BikeType      string                 `protobuf:"bytes,10,opt,name=bike_type,json=bikeType,proto3" json:"bike_type,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *OrderItemResponse) Reset() {
	*x = OrderItemResponse{}
	mi := &file_proto_order_order_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *OrderItemResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderItemResponse) ProtoMessage() {}

func (x *OrderItemResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_order_order_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderItemResponse.ProtoReflect.Descriptor instead.
func (*OrderItemResponse) Descriptor() ([]byte, []int) {
	return file_proto_order_order_proto_rawDescGZIP(), []int{9}
}

func (x *OrderItemResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *OrderItemResponse) GetOrderId() string {
	if x != nil {
		return x.OrderId
	}
	return ""
}

func (x *OrderItemResponse) GetProductId() string {
	if x != nil {
		return x.ProductId
	}
	return ""
}

func (x *OrderItemResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *OrderItemResponse) GetPrice() float64 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *OrderItemResponse) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

func (x *OrderItemResponse) GetFrameSize() string {
	if x != nil {
		return x.FrameSize
	}
	return ""
}

func (x *OrderItemResponse) GetWheelSize() string {
	if x != nil {
		return x.WheelSize
	}
	return ""
}

func (x *OrderItemResponse) GetColor() string {
	if x != nil {
		return x.Color
	}
	return ""
}

func (x *OrderItemResponse) GetBikeType() string {
	if x != nil {
		return x.BikeType
	}
	return ""
}

var File_proto_order_order_proto protoreflect.FileDescriptor

const file_proto_order_order_proto_rawDesc = "" +
	"\n" +
	"\x17proto/order/order.proto\x12\x05order\x1a\x1fgoogle/protobuf/timestamp.proto\"\\\n" +
	"\x12CreateOrderRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12-\n" +
	"\x05items\x18\x02 \x03(\v2\x17.order.OrderItemRequestR\x05items\"\xa0\x02\n" +
	"\rOrderResponse\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\x12*\n" +
	"\x06status\x18\x03 \x01(\x0e2\x12.order.OrderStatusR\x06status\x12\x14\n" +
	"\x05total\x18\x04 \x01(\x01R\x05total\x12.\n" +
	"\x05items\x18\x05 \x03(\v2\x18.order.OrderItemResponseR\x05items\x129\n" +
	"\n" +
	"created_at\x18\a \x01(\v2\x1a.google.protobuf.TimestampR\tcreatedAt\x129\n" +
	"\n" +
	"updated_at\x18\b \x01(\v2\x1a.google.protobuf.TimestampR\tupdatedAt\" \n" +
	"\x0eOrderIDRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\"(\n" +
	"\rUserIDRequest\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\"V\n" +
	"\x18UpdateOrderStatusRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12*\n" +
	"\x06status\x18\x02 \x01(\x0e2\x12.order.OrderStatusR\x06status\"\xf1\x01\n" +
	"\vOrderFilter\x12\x17\n" +
	"\auser_id\x18\x01 \x01(\tR\x06userId\x12*\n" +
	"\x06status\x18\x02 \x01(\x0e2\x12.order.OrderStatusR\x06status\x127\n" +
	"\tfrom_date\x18\x03 \x01(\v2\x1a.google.protobuf.TimestampR\bfromDate\x123\n" +
	"\ato_date\x18\x04 \x01(\v2\x1a.google.protobuf.TimestampR\x06toDate\x12\x12\n" +
	"\x04page\x18\x05 \x01(\x05R\x04page\x12\x1b\n" +
	"\tpage_size\x18\x06 \x01(\x05R\bpageSize\"?\n" +
	"\x11ListOrdersRequest\x12*\n" +
	"\x06filter\x18\x01 \x01(\v2\x12.order.OrderFilterR\x06filter\"\x89\x01\n" +
	"\x12ListOrdersResponse\x12,\n" +
	"\x06orders\x18\x01 \x03(\v2\x14.order.OrderResponseR\x06orders\x12\x14\n" +
	"\x05total\x18\x02 \x01(\x05R\x05total\x12\x12\n" +
	"\x04page\x18\x03 \x01(\x05R\x04page\x12\x1b\n" +
	"\tpage_size\x18\x04 \x01(\x05R\bpageSize\"\xe8\x01\n" +
	"\x10OrderItemRequest\x12\x1d\n" +
	"\n" +
	"product_id\x18\x01 \x01(\tR\tproductId\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x12\x14\n" +
	"\x05price\x18\x03 \x01(\x01R\x05price\x12\x1a\n" +
	"\bquantity\x18\x04 \x01(\x05R\bquantity\x12\x1d\n" +
	"\n" +
	"frame_size\x18\x05 \x01(\tR\tframeSize\x12\x1d\n" +
	"\n" +
	"wheel_size\x18\x06 \x01(\tR\twheelSize\x12\x14\n" +
	"\x05color\x18\a \x01(\tR\x05color\x12\x1b\n" +
	"\tbike_type\x18\b \x01(\tR\bbikeType\"\x94\x02\n" +
	"\x11OrderItemResponse\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x19\n" +
	"\border_id\x18\x02 \x01(\tR\aorderId\x12\x1d\n" +
	"\n" +
	"product_id\x18\x03 \x01(\tR\tproductId\x12\x12\n" +
	"\x04name\x18\x04 \x01(\tR\x04name\x12\x14\n" +
	"\x05price\x18\x05 \x01(\x01R\x05price\x12\x1a\n" +
	"\bquantity\x18\x06 \x01(\x05R\bquantity\x12\x1d\n" +
	"\n" +
	"frame_size\x18\a \x01(\tR\tframeSize\x12\x1d\n" +
	"\n" +
	"wheel_size\x18\b \x01(\tR\twheelSize\x12\x14\n" +
	"\x05color\x18\t \x01(\tR\x05color\x12\x1b\n" +
	"\tbike_type\x18\n" +
	" \x01(\tR\bbikeType*O\n" +
	"\vOrderStatus\x12\v\n" +
	"\aPENDING\x10\x00\x12\b\n" +
	"\x04PAID\x10\x01\x12\v\n" +
	"\aSHIPPED\x10\x02\x12\r\n" +
	"\tDELIVERED\x10\x03\x12\r\n" +
	"\tCANCELLED\x10\x042\xd8\x02\n" +
	"\fOrderService\x12>\n" +
	"\vCreateOrder\x12\x19.order.CreateOrderRequest\x1a\x14.order.OrderResponse\x127\n" +
	"\bGetOrder\x12\x15.order.OrderIDRequest\x1a\x14.order.OrderResponse\x12J\n" +
	"\x11UpdateOrderStatus\x12\x1f.order.UpdateOrderStatusRequest\x1a\x14.order.OrderResponse\x12A\n" +
	"\n" +
	"ListOrders\x12\x18.order.ListOrdersRequest\x1a\x19.order.ListOrdersResponse\x12@\n" +
	"\rGetUserOrders\x12\x14.order.UserIDRequest\x1a\x19.order.ListOrdersResponseB\rZ\vproto/orderb\x06proto3"

var (
	file_proto_order_order_proto_rawDescOnce sync.Once
	file_proto_order_order_proto_rawDescData []byte
)

func file_proto_order_order_proto_rawDescGZIP() []byte {
	file_proto_order_order_proto_rawDescOnce.Do(func() {
		file_proto_order_order_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_order_order_proto_rawDesc), len(file_proto_order_order_proto_rawDesc)))
	})
	return file_proto_order_order_proto_rawDescData
}

var file_proto_order_order_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_order_order_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_proto_order_order_proto_goTypes = []any{
	(OrderStatus)(0),                 // 0: order.OrderStatus
	(*CreateOrderRequest)(nil),       // 1: order.CreateOrderRequest
	(*OrderResponse)(nil),            // 2: order.OrderResponse
	(*OrderIDRequest)(nil),           // 3: order.OrderIDRequest
	(*UserIDRequest)(nil),            // 4: order.UserIDRequest
	(*UpdateOrderStatusRequest)(nil), // 5: order.UpdateOrderStatusRequest
	(*OrderFilter)(nil),              // 6: order.OrderFilter
	(*ListOrdersRequest)(nil),        // 7: order.ListOrdersRequest
	(*ListOrdersResponse)(nil),       // 8: order.ListOrdersResponse
	(*OrderItemRequest)(nil),         // 9: order.OrderItemRequest
	(*OrderItemResponse)(nil),        // 10: order.OrderItemResponse
	(*timestamppb.Timestamp)(nil),    // 11: google.protobuf.Timestamp
}
var file_proto_order_order_proto_depIdxs = []int32{
	9,  // 0: order.CreateOrderRequest.items:type_name -> order.OrderItemRequest
	0,  // 1: order.OrderResponse.status:type_name -> order.OrderStatus
	10, // 2: order.OrderResponse.items:type_name -> order.OrderItemResponse
	11, // 3: order.OrderResponse.created_at:type_name -> google.protobuf.Timestamp
	11, // 4: order.OrderResponse.updated_at:type_name -> google.protobuf.Timestamp
	0,  // 5: order.UpdateOrderStatusRequest.status:type_name -> order.OrderStatus
	0,  // 6: order.OrderFilter.status:type_name -> order.OrderStatus
	11, // 7: order.OrderFilter.from_date:type_name -> google.protobuf.Timestamp
	11, // 8: order.OrderFilter.to_date:type_name -> google.protobuf.Timestamp
	6,  // 9: order.ListOrdersRequest.filter:type_name -> order.OrderFilter
	2,  // 10: order.ListOrdersResponse.orders:type_name -> order.OrderResponse
	1,  // 11: order.OrderService.CreateOrder:input_type -> order.CreateOrderRequest
	3,  // 12: order.OrderService.GetOrder:input_type -> order.OrderIDRequest
	5,  // 13: order.OrderService.UpdateOrderStatus:input_type -> order.UpdateOrderStatusRequest
	7,  // 14: order.OrderService.ListOrders:input_type -> order.ListOrdersRequest
	4,  // 15: order.OrderService.GetUserOrders:input_type -> order.UserIDRequest
	2,  // 16: order.OrderService.CreateOrder:output_type -> order.OrderResponse
	2,  // 17: order.OrderService.GetOrder:output_type -> order.OrderResponse
	2,  // 18: order.OrderService.UpdateOrderStatus:output_type -> order.OrderResponse
	8,  // 19: order.OrderService.ListOrders:output_type -> order.ListOrdersResponse
	8,  // 20: order.OrderService.GetUserOrders:output_type -> order.ListOrdersResponse
	16, // [16:21] is the sub-list for method output_type
	11, // [11:16] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_proto_order_order_proto_init() }
func file_proto_order_order_proto_init() {
	if File_proto_order_order_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_order_order_proto_rawDesc), len(file_proto_order_order_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_order_order_proto_goTypes,
		DependencyIndexes: file_proto_order_order_proto_depIdxs,
		EnumInfos:         file_proto_order_order_proto_enumTypes,
		MessageInfos:      file_proto_order_order_proto_msgTypes,
	}.Build()
	File_proto_order_order_proto = out.File
	file_proto_order_order_proto_goTypes = nil
	file_proto_order_order_proto_depIdxs = nil
}
