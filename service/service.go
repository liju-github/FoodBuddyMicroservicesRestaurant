package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	model "github.com/liju-github/FoodBuddyMicroserviceRestaurant/models"
	"github.com/liju-github/FoodBuddyMicroserviceRestaurant/repository"
)

type RestaurantService struct {
	restaurantPb.UnimplementedRestaurantServiceServer
	repo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) *RestaurantService {
	return &RestaurantService{
		repo: repo,
	}
}

func (s *RestaurantService) RestaurantSignup(ctx context.Context, req *restaurantPb.RestaurantSignupRequest) (*restaurantPb.RestaurantSignupResponse, error) {
	// Check if email already exists
	existingRestaurant, err := s.repo.GetRestaurantByEmail(req.OwnerEmail)
	if err == nil && existingRestaurant != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create restaurant
	restaurant := &model.Restaurant{
		ID:           fmt.Sprintf("rest_%s", uuid.New().String()),
		OwnerEmail:   req.OwnerEmail,
		PasswordHash: string(hashedPassword),
		Name:         req.RestaurantName,
		PhoneNumber:  req.PhoneNumber,
		StreetName:   req.Address.StreetName,
		Locality:     req.Address.Locality,
		State:        req.Address.State,
		Pincode:      req.Address.Pincode,
	}

	if err := s.repo.CreateRestaurant(restaurant); err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %v", err)
	}

	return &restaurantPb.RestaurantSignupResponse{
		RestaurantId: restaurant.ID,
		Message:      "Restaurant registered successfully",
	}, nil
}

func (s *RestaurantService) RestaurantLogin(ctx context.Context, req *restaurantPb.RestaurantLoginRequest) (*restaurantPb.RestaurantLoginResponse, error) {
	restaurant, err := s.repo.GetRestaurantByEmail(req.OwnerEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(restaurant.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if restaurant.IsBanned {
		return nil, fmt.Errorf("restaurant is banned: %s", restaurant.BanReason)
	}

	return &restaurantPb.RestaurantLoginResponse{
		RestaurantId: restaurant.ID,
		Message:      "Login successful",
	}, nil
}

func (s *RestaurantService) EditRestaurant(ctx context.Context, req *restaurantPb.EditRestaurantRequest) (*restaurantPb.EditRestaurantResponse, error) {

	restaurant, err := s.repo.GetRestaurantByID(req.RestaurantId)
	if err != nil {
		return nil, err
	}

	restaurant.Name = req.RestaurantName
	restaurant.PhoneNumber = req.PhoneNumber
	restaurant.StreetName = req.Address.StreetName
	restaurant.Locality = req.Address.Locality
	restaurant.State = req.Address.State
	restaurant.Pincode = req.Address.Pincode

	if err := s.repo.UpdateRestaurant(restaurant); err != nil {
		return nil, fmt.Errorf("failed to update restaurant: %v", err)
	}

	return &restaurantPb.EditRestaurantResponse{
		Message: "Restaurant updated successfully",
	}, nil
}

func (s *RestaurantService) GetRestaurantProductsByID(ctx context.Context, req *restaurantPb.GetRestaurantProductsByIDRequest) (*restaurantPb.GetRestaurantProductsByIDResponse, error) {
	products, err := s.repo.GetProductsByRestaurantID(req.RestaurantId)
	if err != nil {
		return nil, err
	}

	var pbProducts []*restaurantPb.Product
	for _, p := range products {
		pbProducts = append(pbProducts, &restaurantPb.Product{
			ProductId:    p.ID,
			RestaurantId: p.RestaurantID,
			Name:         p.Name,
			Description:  p.Description,
			Price:        p.Price,
			Stock:        p.Stock,
			Category:     p.Category,
		})
	}

	return &restaurantPb.GetRestaurantProductsByIDResponse{
		Products: pbProducts,
		Message:  "Products retrieved successfully",
	}, nil
}

func (s *RestaurantService) GetAllRestaurantWithProducts(ctx context.Context, req *restaurantPb.GetAllRestaurantAndProductsRequest) (*restaurantPb.GetAllRestaurantWithProductsResponse, error) {
	restaurants, err := s.repo.GetAllRestaurants()
	if err != nil {
		return nil, err
	}

	var pbRestaurants []*restaurantPb.RestaurantWithProducts
	for _, r := range restaurants {
		products, err := s.repo.GetProductsByRestaurantID(r.ID)
		if err != nil {
			continue
		}

		var pbProducts []*restaurantPb.Product
		for _, p := range products {
			pbProducts = append(pbProducts, &restaurantPb.Product{
				ProductId:    p.ID,
				RestaurantId: p.RestaurantID,
				Name:         p.Name,
				Description:  p.Description,
				Price:        p.Price,
				Stock:        p.Stock,
				Category:     p.Category,
			})
		}

		pbRestaurants = append(pbRestaurants, &restaurantPb.RestaurantWithProducts{
			RestaurantId:   r.ID,
			RestaurantName: r.Name,
			PhoneNumber:    r.PhoneNumber,
			Address: &restaurantPb.Address{
				StreetName: r.StreetName,
				Locality:   r.Locality,
				State:      r.State,
				Pincode:    r.Pincode,
			},
			Products: pbProducts,
		})
	}

	return &restaurantPb.GetAllRestaurantWithProductsResponse{
		Restaurants: pbRestaurants,
		Message:     "Restaurants with products retrieved successfully",
	}, nil
}

func (s *RestaurantService) AddProduct(ctx context.Context, req *restaurantPb.AddProductRequest) (*restaurantPb.AddProductResponse, error) {

	product := &model.Product{
		ID:           fmt.Sprintf("prod_%s", uuid.New().String()),
		RestaurantID: req.RestaurantId,
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		Stock:        req.Stock,
		Category:     req.Category,
	}

	if err := s.repo.AddProduct(product); err != nil {
		return nil, fmt.Errorf("failed to add product: %v", err)
	}

	return &restaurantPb.AddProductResponse{
		ProductId: product.ID,
		Message:   "Product added successfully",
	}, nil
}

func (s *RestaurantService) EditProduct(ctx context.Context, req *restaurantPb.EditProductRequest) (*restaurantPb.EditProductResponse, error) {

	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.Category = req.Category
	product.RestaurantID = req.RestaurantId

	if err := s.repo.UpdateProduct(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %v", err)
	}

	return &restaurantPb.EditProductResponse{
		Message: "Product updated successfully",
	}, nil
}

func (s *RestaurantService) GetProductByID(ctx context.Context, req *restaurantPb.GetProductByIDRequest) (*restaurantPb.GetProductByIDResponse, error) {
	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	return &restaurantPb.GetProductByIDResponse{
		Product: &restaurantPb.Product{
			ProductId:    product.ID,
			RestaurantId: product.RestaurantID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Stock:        product.Stock,
			Category:     product.Category,
		},
		Message: "Product retrieved successfully",
	}, nil
}

func (s *RestaurantService) DeleteProductByID(ctx context.Context, req *restaurantPb.DeleteProductByIDRequest) (*restaurantPb.DeleteProductByIDResponse, error) {

	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	product.RestaurantID = req.RestaurantId

	if err := s.repo.DeleteProduct(req.ProductId); err != nil {
		return nil, err
	}

	return &restaurantPb.DeleteProductByIDResponse{
		Message: "Product deleted successfully",
	}, nil
}

func (s *RestaurantService) IncremenentProductStockByValue(ctx context.Context, req *restaurantPb.IncremenentProductStockByValueRequest) (*restaurantPb.IncremenentProductStockByValueResponse, error) {

	if err := s.repo.UpdateProductStock(req.ProductId, req.Value); err != nil {
		return nil, err
	}

	return &restaurantPb.IncremenentProductStockByValueResponse{
		Message: "Stock incremented successfully",
	}, nil
}

func (s *RestaurantService) DecrementProductStockByValue(ctx context.Context, req *restaurantPb.DecrementProductStockByValueByValueRequest) (*restaurantPb.DecrementProductStockByValueResponse, error) {

	currentStock, err := s.repo.GetProductStock(req.ProductId)
	if err != nil {
		return nil, err
	}

	if currentStock < req.Value {
		return nil, fmt.Errorf("insufficient stock")
	}

	if err := s.repo.UpdateProductStock(req.ProductId, -req.Value); err != nil {
		return nil, err
	}

	return &restaurantPb.DecrementProductStockByValueResponse{
		Message: "Stock decremented successfully",
	}, nil
}

func (s *RestaurantService) GetRestaurantIDviaProductID(ctx context.Context, req *restaurantPb.GetRestaurantIDviaProductIDRequest) (*restaurantPb.GetRestaurantIDviaProductIDResponse, error) {
	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	return &restaurantPb.GetRestaurantIDviaProductIDResponse{
		RestaurantId: product.RestaurantID,
		Message:      "Restaurant ID retrieved successfully",
	}, nil
}

func (s *RestaurantService) GetStockByProductID(ctx context.Context, req *restaurantPb.GetStockByProductIDRequest) (*restaurantPb.GetStockByProductIDResponse, error) {
	stock, err := s.repo.GetProductStock(req.ProductId)
	if err != nil {
		return nil, err
	}

	return &restaurantPb.GetStockByProductIDResponse{
		Stock:   stock,
		Message: "Stock retrieved successfully",
	}, nil
}

func (s *RestaurantService) BanRestaurant(ctx context.Context, req *restaurantPb.BanRestaurantRequest) (*restaurantPb.BanRestaurantResponse, error) {
	if err := s.repo.BanRestaurant(req.RestaurantId, req.Reason); err != nil {
		return nil, err
	}

	return &restaurantPb.BanRestaurantResponse{
		Message: "Restaurant banned successfully",
	}, nil
}

func (s *RestaurantService) UnbanRestaurant(ctx context.Context, req *restaurantPb.UnbanRestaurantRequest) (*restaurantPb.UnbanRestaurantResponse, error) {
	if err := s.repo.UnbanRestaurant(req.RestaurantId); err != nil {
		return nil, err
	}

	return &restaurantPb.UnbanRestaurantResponse{
		Message: "Restaurant unbanned successfully",
	}, nil
}

func (s *RestaurantService) GetAllProducts(ctx context.Context, req *restaurantPb.GetAllProductsRequest) (*restaurantPb.GetAllProductsResponse, error) {
	products, err := s.repo.GetAllProducts()
	if err != nil {
		return nil, err
	}

	var pbProducts []*restaurantPb.Product
	for _, product := range products {
		pbProducts = append(pbProducts, &restaurantPb.Product{
			ProductId:    product.ID,
			RestaurantId: product.RestaurantID,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Stock:        product.Stock,
			Category:     product.Category,
		})
	}

	return &restaurantPb.GetAllProductsResponse{
		Products: pbProducts,
		Message:  "All products retrieved successfully",
	}, nil
}

func (s *RestaurantService) CheckRestaurantBanStatus(ctx context.Context, req *restaurantPb.CheckRestaurantBanStatusRequest) (*restaurantPb.CheckRestaurantBanStatusResponse, error) {
	restaurant, err := s.repo.GetRestaurantByID(req.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get restaurant: %v", err)
	}

	if restaurant == nil {
		return &restaurantPb.CheckRestaurantBanStatusResponse{
			IsBanned: false,
			Message:  "Restaurant not found",
		}, nil
	}

	return &restaurantPb.CheckRestaurantBanStatusResponse{
		IsBanned: restaurant.IsBanned,
		Reason:   restaurant.BanReason,
		Message:  "Ban status retrieved successfully",
	}, nil
}

func (s *RestaurantService) GetRestaurantByID(ctx context.Context, req *restaurantPb.GetRestaurantByIDRequest) (*restaurantPb.GetRestaurantByIDResponse, error) {
	restaurant, err := s.repo.GetRestaurantByID(req.RestaurantId)
	if err != nil {
		return &restaurantPb.GetRestaurantByIDResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get restaurant: %v", err),
		}, nil
	}

	return &restaurantPb.GetRestaurantByIDResponse{
		Success:        true,
		Message:        "Restaurant found successfully",
		RestaurantId:   restaurant.ID,
		RestaurantName: restaurant.Name,
		PhoneNumber:    restaurant.PhoneNumber,
		Address: &restaurantPb.Address{
			StreetName: restaurant.StreetName,
			Locality:   restaurant.Locality,
			State:      restaurant.State,
			Pincode:    restaurant.Pincode,
		},
		IsBanned:  restaurant.IsBanned,
		BanReason: restaurant.BanReason,
	}, nil
}
