package handler

import (
	"context"
	"log"

	"inventory-service/internal/domain"
	"inventory-service/internal/service"

	pb "proto/inventory"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductGrpcHandler struct {
	pb.UnimplementedProductServiceServer
	productService service.ProductService
}

func NewProductGrpcHandler(productService service.ProductService) *ProductGrpcHandler {
	return &ProductGrpcHandler{
		productService: productService,
	}
}

func (h *ProductGrpcHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	log.Printf("Received CreateProduct request for %s", req.Name)

	product := domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		CategoryID:  req.CategoryId,
	}

	createdProduct, err := h.productService.CreateProduct(ctx, product)
	if err != nil {
		log.Printf("Failed to create product: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &pb.ProductResponse{
		Id:          createdProduct.ID,
		Name:        createdProduct.Name,
		Description: createdProduct.Description,
		Price:       createdProduct.Price,
		Stock:       int32(createdProduct.Stock),
		CategoryId:  createdProduct.CategoryID,
		CreatedAt:   timestamppb.New(createdProduct.CreatedAt),
		UpdatedAt:   timestamppb.New(createdProduct.UpdatedAt),
	}, nil
}

func (h *ProductGrpcHandler) GetProduct(ctx context.Context, req *pb.ProductIDRequest) (*pb.ProductResponse, error) {
	log.Printf("Received GetProduct request for ID: %s", req.Id)

	product, err := h.productService.GetProductByID(ctx, req.Id)
	if err != nil {
		log.Printf("Failed to get product: %v", err)
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	return &pb.ProductResponse{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int32(product.Stock),
		CategoryId:  product.CategoryID,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}, nil
}

func (h *ProductGrpcHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	log.Printf("Received UpdateProduct request for ID: %s", req.Id)

	product := domain.Product{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		CategoryID:  req.CategoryId,
	}

	if err := h.productService.UpdateProduct(ctx, product); err != nil {
		log.Printf("Failed to update product: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	// Get the updated product
	updatedProduct, err := h.productService.GetProductByID(ctx, req.Id)
	if err != nil {
		log.Printf("Failed to get updated product: %v", err)
		return nil, status.Errorf(codes.NotFound, "failed to get updated product: %v", err)
	}

	return &pb.ProductResponse{
		Id:          updatedProduct.ID,
		Name:        updatedProduct.Name,
		Description: updatedProduct.Description,
		Price:       updatedProduct.Price,
		Stock:       int32(updatedProduct.Stock),
		CategoryId:  updatedProduct.CategoryID,
		CreatedAt:   timestamppb.New(updatedProduct.CreatedAt),
		UpdatedAt:   timestamppb.New(updatedProduct.UpdatedAt),
	}, nil
}

func (h *ProductGrpcHandler) DeleteProduct(ctx context.Context, req *pb.ProductIDRequest) (*pb.DeleteResponse, error) {
	log.Printf("Received DeleteProduct request for ID: %s", req.Id)

	if err := h.productService.DeleteProduct(ctx, req.Id); err != nil {
		log.Printf("Failed to delete product: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Product deleted successfully",
	}, nil
}

func (h *ProductGrpcHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	log.Printf("Received ListProducts request")

	filter := domain.ProductFilter{
		CategoryID: req.Filter.CategoryId,
		Page:       int(req.Filter.Page),
		PageSize:   int(req.Filter.PageSize),
	}

	if req.Filter.MinPrice > 0 {
		minPrice := req.Filter.MinPrice
		filter.MinPrice = &minPrice
	}

	if req.Filter.MaxPrice > 0 {
		maxPrice := req.Filter.MaxPrice
		filter.MaxPrice = &maxPrice
	}

	if req.Filter.InStock {
		inStock := req.Filter.InStock
		filter.InStock = &inStock
	}

	products, total, err := h.productService.ListProducts(ctx, filter)
	if err != nil {
		log.Printf("Failed to list products: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	var protoProducts []*pb.ProductResponse
	for _, product := range products {
		protoProducts = append(protoProducts, &pb.ProductResponse{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       int32(product.Stock),
			CategoryId:  product.CategoryID,
			CreatedAt:   timestamppb.New(product.CreatedAt),
			UpdatedAt:   timestamppb.New(product.UpdatedAt),
		})
	}

	return &pb.ListProductsResponse{
		Products: protoProducts,
		Total:    int32(total),
		Page:     int32(filter.Page),
		PageSize: int32(filter.PageSize),
	}, nil
}

func (h *ProductGrpcHandler) CheckStock(ctx context.Context, req *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	log.Printf("Received CheckStock request for %d items", len(req.Items))

	// Get all requested products
	var unavailableItems []*pb.ProductQuantity
	for _, item := range req.Items {
		product, err := h.productService.GetProductByID(ctx, item.ProductId)
		if err != nil {
			log.Printf("Product %s not found: %v", item.ProductId, err)
			unavailableItems = append(unavailableItems, &pb.ProductQuantity{
				ProductId: item.ProductId,
				Quantity:  item.Quantity,
			})
			continue
		}

		if product.Stock < int(item.Quantity) {
			log.Printf("Insufficient stock for product %s: requested %d, available %d",
				item.ProductId, item.Quantity, product.Stock)
			unavailableItems = append(unavailableItems, &pb.ProductQuantity{
				ProductId: item.ProductId,
				Quantity:  item.Quantity,
			})
		}
	}

	available := len(unavailableItems) == 0
	return &pb.CheckStockResponse{
		Available:        available,
		UnavailableItems: unavailableItems,
	}, nil
}

type CategoryGrpcHandler struct {
	pb.UnimplementedCategoryServiceServer
	categoryService service.CategoryService
}

func NewCategoryGrpcHandler(categoryService service.CategoryService) *CategoryGrpcHandler {
	return &CategoryGrpcHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryGrpcHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	log.Printf("Received CreateCategory request for %s", req.Name)

	category := domain.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	createdCategory, err := h.categoryService.CreateCategory(ctx, category)
	if err != nil {
		log.Printf("Failed to create category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	return &pb.CategoryResponse{
		Id:          createdCategory.ID,
		Name:        createdCategory.Name,
		Description: createdCategory.Description,
		CreatedAt:   timestamppb.New(createdCategory.CreatedAt),
		UpdatedAt:   timestamppb.New(createdCategory.UpdatedAt),
	}, nil
}

func (h *CategoryGrpcHandler) GetCategory(ctx context.Context, req *pb.CategoryIDRequest) (*pb.CategoryResponse, error) {
	log.Printf("Received GetCategory request for ID: %s", req.Id)

	category, err := h.categoryService.GetCategoryByID(ctx, req.Id)
	if err != nil {
		log.Printf("Failed to get category: %v", err)
		return nil, status.Errorf(codes.NotFound, "category not found: %v", err)
	}

	return &pb.CategoryResponse{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   timestamppb.New(category.CreatedAt),
		UpdatedAt:   timestamppb.New(category.UpdatedAt),
	}, nil
}

func (h *CategoryGrpcHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.CategoryResponse, error) {
	log.Printf("Received UpdateCategory request for ID: %s", req.Id)

	category := domain.Category{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.categoryService.UpdateCategory(ctx, category); err != nil {
		log.Printf("Failed to update category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}

	// Get the updated category
	updatedCategory, err := h.categoryService.GetCategoryByID(ctx, req.Id)
	if err != nil {
		log.Printf("Failed to get updated category: %v", err)
		return nil, status.Errorf(codes.NotFound, "failed to get updated category: %v", err)
	}

	return &pb.CategoryResponse{
		Id:          updatedCategory.ID,
		Name:        updatedCategory.Name,
		Description: updatedCategory.Description,
		CreatedAt:   timestamppb.New(updatedCategory.CreatedAt),
		UpdatedAt:   timestamppb.New(updatedCategory.UpdatedAt),
	}, nil
}

func (h *CategoryGrpcHandler) DeleteCategory(ctx context.Context, req *pb.CategoryIDRequest) (*pb.DeleteCategoryResponse, error) {
	log.Printf("Received DeleteCategory request for ID: %s", req.Id)

	if err := h.categoryService.DeleteCategory(ctx, req.Id); err != nil {
		log.Printf("Failed to delete category: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}

	return &pb.DeleteCategoryResponse{
		Success: true,
		Message: "Category deleted successfully",
	}, nil
}

func (h *CategoryGrpcHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	log.Printf("Received ListCategories request")

	categories, err := h.categoryService.ListCategories(ctx)
	if err != nil {
		log.Printf("Failed to list categories: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	var protoCategories []*pb.CategoryResponse
	for _, category := range categories {
		protoCategories = append(protoCategories, &pb.CategoryResponse{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   timestamppb.New(category.CreatedAt),
			UpdatedAt:   timestamppb.New(category.UpdatedAt),
		})
	}

	return &pb.ListCategoriesResponse{
		Categories: protoCategories,
	}, nil
}
