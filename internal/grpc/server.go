package grpc

import (
	"context"

	pb "github.com/savanrajyaguru/ecommerce-go-inventory-service/proto"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/services"
)

type InventoryGrpcServer struct {
	pb.UnimplementedInventoryServiceServer
	service *services.InventoryService
}

func NewInventoryGrpcServer(service *services.InventoryService) *InventoryGrpcServer {
	return &InventoryGrpcServer{service: service}
}

func (s *InventoryGrpcServer) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	err := s.service.ReserveStock(ctx, uint(req.ProductId), int(req.Quantity))
	if err != nil {
		return &pb.ReserveStockResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ReserveStockResponse{
		Success: true,
		Message: "Stock reserved successfully",
	}, nil
}
