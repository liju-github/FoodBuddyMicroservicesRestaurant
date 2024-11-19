package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	config "github.com/liju-github/FoodBuddyMicroserviceRestaurant/configs"
	model "github.com/liju-github/FoodBuddyMicroserviceRestaurant/models"
	"github.com/liju-github/FoodBuddyMicroserviceRestaurant/repository"
)

const (
	TokenExpiry = 24 * time.Hour
)

type RestaurantService struct {
	restaurantPb.UnimplementedRestaurantServiceServer
	repo      repository.RestaurantRepository
	JWTSecret string
}

func NewRestaurantService(repo repository.RestaurantRepository) *RestaurantService {
	return &RestaurantService{
		repo:      repo,
		JWTSecret: config.LoadConfig().JWTSecretKey,
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
		ID:           uuid.New().String(),
		OwnerEmail:   req.OwnerEmail,
		PasswordHash: string(hashedPassword),
		Name:         req.RestaurantName,
		PhoneNumber:  req.PhoneNumber,
		Address: &model.Address{
			ID:         uuid.New().String(),
			StreetName: req.Address.StreetName,
			Locality:   req.Address.Locality,
			State:      req.Address.State,
			Pincode:    req.Address.Pincode,
		},
	}

	if err := s.repo.CreateRestaurant(restaurant); err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %v", err)
	}

	// Generate JWT token
	token, err := s.generateToken(restaurant.ID, restaurant.OwnerEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &restaurantPb.RestaurantSignupResponse{
		RestaurantId: restaurant.ID,
		Token:        token,
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

	token, err := s.generateToken(restaurant.ID, restaurant.OwnerEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &restaurantPb.RestaurantLoginResponse{
		RestaurantId: restaurant.ID,
		Token:        token,
		Message:      "Login successful",
	}, nil
}

func (s *RestaurantService) EditRestaurant(ctx context.Context, req *restaurantPb.EditRestaurantRequest) (*restaurantPb.EditRestaurantResponse, error) {
	// Get restaurant ID from context
	restaurantID := ctx.Value("restaurant_id").(string)
	if restaurantID != req.RestaurantId {
		return nil, fmt.Errorf("unauthorized: you can only edit your own restaurant")
	}

	restaurant, err := s.repo.GetRestaurantByID(req.RestaurantId)
	if err != nil {
		return nil, err
	}

	restaurant.Name = req.RestaurantName
	restaurant.PhoneNumber = req.PhoneNumber
	restaurant.Address = &model.Address{
		ID:         uuid.New().String(),
		StreetName: req.Address.StreetName,
		Locality:   req.Address.Locality,
		State:      req.Address.State,
		Pincode:    req.Address.Pincode,
	}

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
				AddressId:  r.Address.ID,
				StreetName: r.Address.StreetName,
				Locality:   r.Address.Locality,
				State:      r.Address.State,
				Pincode:    r.Address.Pincode,
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
	// Get restaurant ID from context
	restaurantID := ctx.Value("restaurant_id").(string)
	if restaurantID != req.RestaurantId {
		return nil, fmt.Errorf("unauthorized: you can only add products to your own restaurant")
	}

	product := &model.Product{
		ID:           uuid.New().String(),
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
	// Get restaurant ID from context
	restaurantID := ctx.Value("restaurant_id").(string)

	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	if restaurantID != product.RestaurantID {
		return nil, fmt.Errorf("unauthorized: you can only edit your own products")
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.Category = req.Category

	if err := s.repo.UpdateProduct(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %v", err)
	}

	return &restaurantPb.EditProductResponse{
		Message: "Product updated successfully",
	}, nil
}

func (s *RestaurantService) DeleteProductByID(ctx context.Context, req *restaurantPb.DeleteProductByIDRequest) (*restaurantPb.DeleteProductByIDResponse, error) {
	// Get restaurant ID from context
	restaurantID := ctx.Value("restaurant_id").(string)

	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	if restaurantID != product.RestaurantID {
		return nil, fmt.Errorf("unauthorized: you can only delete your own products")
	}

	if err := s.repo.DeleteProduct(req.ProductId); err != nil {
		return nil, err
	}

	return &restaurantPb.DeleteProductByIDResponse{
		Message: "Product deleted successfully",
	}, nil
}

func (s *RestaurantService) IncremenentProductStockByValue(ctx context.Context, req *restaurantPb.IncremenentProductStockByValueRequest) (*restaurantPb.IncremenentProductStockByValueResponse, error) {
	// Get restaurant ID from context
	restaurantID := ctx.Value("restaurant_id").(string)

	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	if restaurantID != product.RestaurantID {
		return nil, fmt.Errorf("unauthorized: you can only modify stock of your own products")
	}

	if err := s.repo.UpdateProductStock(req.ProductId, req.Value); err != nil {
		return nil, err
	}

	return &restaurantPb.IncremenentProductStockByValueResponse{
		Message: "Stock incremented successfully",
	}, nil
}

func (s *RestaurantService) DecrementProductStockByValue(ctx context.Context, req *restaurantPb.DecrementProductStockByValueByValueRequest) (*restaurantPb.DecrementProductStockByValueResponse, error) {
	// Get restaurant ID from context
	restaurantID := ctx.Value("restaurant_id").(string)

	product, err := s.repo.GetProductByID(req.ProductId)
	if err != nil {
		return nil, err
	}

	if restaurantID != product.RestaurantID {
		return nil, fmt.Errorf("unauthorized: you can only modify stock of your own products")
	}

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

func (s *RestaurantService) generateToken(restaurantID, email string) (string, error) {
	claims := jwt.MapClaims{
		"id":    restaurantID,
		"email": email,
		"exp":   time.Now().Add(TokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}
