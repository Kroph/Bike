package handler

import (
	"context"
	"log"

	pb "proto/order"

	"order-service/internal/domain"
	"order-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGrpcHandler struct {
	pb.UnimplementedOrderServiceServer
	orderService service.OrderService
}

func NewOrderGrpcHandler(orderService service.OrderService) *OrderGrpcHandler {
	return &OrderGrpcHandler{
		orderService: orderService,
	}
}

func (h *OrderGrpcHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	log.Printf("Received CreateOrder request for user: %s", req.UserId)

	var orderItems []domain.OrderItem
	for _, item := range req.Items {
		orderItems = append(orderItems, domain.OrderItem{
			ProductID: item.ProductId,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  int(item.Quantity),
			FrameSize: item.FrameSize,
			WheelSize: item.WheelSize,
			Color:     item.Color,
			BikeType:  item.BikeType,
		})
	}

	order := domain.Order{
		UserID: req.UserId,
		Items:  orderItems,
		Status: domain.OrderStatusPending,
	}

	createdOrder, err := h.orderService.CreateOrder(ctx, order)
	if err != nil {
		log.Printf("Failed to create order: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return mapOrderToProto(createdOrder), nil
}

func (h *OrderGrpcHandler) GetOrder(ctx context.Context, req *pb.OrderIDRequest) (*pb.OrderResponse, error) {
	log.Printf("Received GetOrder request for ID: %s", req.Id)

	order, err := h.orderService.GetOrderByID(ctx, req.Id)
	if err != nil {
		log.Printf("Failed to get order: %v", err)
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	return mapOrderToProto(order), nil
}

func (h *OrderGrpcHandler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderResponse, error) {
	log.Printf("Received UpdateOrderStatus request for ID: %s", req.Id)

	var orderStatus domain.OrderStatus
	switch req.Status {
	case pb.OrderStatus_PENDING:
		orderStatus = domain.OrderStatusPending
	case pb.OrderStatus_PAID:
		orderStatus = domain.OrderStatusPaid
	case pb.OrderStatus_SHIPPED:
		orderStatus = domain.OrderStatusShipped
	case pb.OrderStatus_DELIVERED:
		orderStatus = domain.OrderStatusDelivered
	case pb.OrderStatus_CANCELLED:
		orderStatus = domain.OrderStatusCancelled
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid order status")
	}

	if err := h.orderService.UpdateOrderStatus(ctx, req.Id, orderStatus); err != nil {
		log.Printf("Failed to update order status: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update order status: %v", err)
	}

	// Get the updated order
	updatedOrder, err := h.orderService.GetOrderByID(ctx, req.Id)
	if err != nil {
		log.Printf("Failed to get updated order: %v", err)
		return nil, status.Errorf(codes.NotFound, "failed to get updated order: %v", err)
	}

	return mapOrderToProto(updatedOrder), nil
}

func (h *OrderGrpcHandler) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	log.Printf("Received ListOrders request")

	filter := domain.OrderFilter{
		UserID:   req.Filter.UserId,
		Page:     int(req.Filter.Page),
		PageSize: int(req.Filter.PageSize),
	}

	switch req.Filter.Status {
	case pb.OrderStatus_PENDING:
		filter.Status = domain.OrderStatusPending
	case pb.OrderStatus_PAID:
		filter.Status = domain.OrderStatusPaid
	case pb.OrderStatus_SHIPPED:
		filter.Status = domain.OrderStatusShipped
	case pb.OrderStatus_DELIVERED:
		filter.Status = domain.OrderStatusDelivered
	case pb.OrderStatus_CANCELLED:
		filter.Status = domain.OrderStatusCancelled
	}

	if req.Filter.FromDate != nil {
		fromDate := req.Filter.FromDate.AsTime()
		filter.FromDate = &fromDate
	}

	if req.Filter.ToDate != nil {
		toDate := req.Filter.ToDate.AsTime()
		filter.ToDate = &toDate
	}

	orders, total, err := h.orderService.ListOrders(ctx, filter)
	if err != nil {
		log.Printf("Failed to list orders: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list orders: %v", err)
	}

	var protoOrders []*pb.OrderResponse
	for _, order := range orders {
		protoOrders = append(protoOrders, mapOrderToProto(order))
	}

	return &pb.ListOrdersResponse{
		Orders:   protoOrders,
		Total:    int32(total),
		Page:     int32(filter.Page),
		PageSize: int32(filter.PageSize),
	}, nil
}

func (h *OrderGrpcHandler) GetUserOrders(ctx context.Context, req *pb.UserIDRequest) (*pb.ListOrdersResponse, error) {
	log.Printf("Received GetUserOrders request for user: %s", req.UserId)

	orders, err := h.orderService.GetUserOrders(ctx, req.UserId)
	if err != nil {
		log.Printf("Failed to get user orders: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get user orders: %v", err)
	}

	var protoOrders []*pb.OrderResponse
	for _, order := range orders {
		protoOrders = append(protoOrders, mapOrderToProto(order))
	}

	return &pb.ListOrdersResponse{
		Orders: protoOrders,
		Total:  int32(len(orders)),
		Page:   1,
	}, nil
}

// Helper function to map domain.Order to pb.OrderResponse
func mapOrderToProto(order domain.Order) *pb.OrderResponse {
	var status pb.OrderStatus
	switch order.Status {
	case domain.OrderStatusPending:
		status = pb.OrderStatus_PENDING
	case domain.OrderStatusPaid:
		status = pb.OrderStatus_PAID
	case domain.OrderStatusShipped:
		status = pb.OrderStatus_SHIPPED
	case domain.OrderStatusDelivered:
		status = pb.OrderStatus_DELIVERED
	case domain.OrderStatusCancelled:
		status = pb.OrderStatus_CANCELLED
	}

	var items []*pb.OrderItemResponse
	for _, item := range order.Items {
		items = append(items, &pb.OrderItemResponse{
			Id:        item.ID,
			OrderId:   item.OrderID,
			ProductId: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  int32(item.Quantity),
			FrameSize: item.FrameSize,
			WheelSize: item.WheelSize,
			Color:     item.Color,
			BikeType:  item.BikeType,
		})
	}

	return &pb.OrderResponse{
		Id:        order.ID,
		UserId:    order.UserID,
		Status:    status,
		Total:     order.Total,
		Items:     items,
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}
}
