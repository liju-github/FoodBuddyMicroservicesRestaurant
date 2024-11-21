package repository

import (
	"errors"
	"fmt"

	model "github.com/liju-github/FoodBuddyMicroserviceRestaurant/models"
	"gorm.io/gorm"
)

type RestaurantRepository interface {
	CreateRestaurant(restaurant *model.Restaurant) error
	GetRestaurantByEmail(email string) (*model.Restaurant, error)
	GetRestaurantByID(id string) (*model.Restaurant, error)
	UpdateRestaurant(restaurant *model.Restaurant) error
	GetAllRestaurants() ([]*model.Restaurant, error)
	BanRestaurant(restaurantID, reason string) error
	UnbanRestaurant(restaurantID string) error

	AddProduct(product *model.Product) error
	GetProductByID(productID string) (*model.Product, error)
	GetProductsByRestaurantID(restaurantID string) ([]*model.Product, error)
	UpdateProduct(product *model.Product) error
	DeleteProduct(productID string) error
	GetAllProducts() ([]*model.Product, error)
	UpdateProductStock(productID string, quantity int32) error
	GetProductStock(productID string) (int32, error)
	GetRestaurantWithProducts(restaurantID string) (*model.Restaurant, []*model.Product, error)
	GetAllRestaurantsWithProducts() ([]*model.Restaurant, error)
}

type restaurantRepository struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &restaurantRepository{db: db}
}

// Restaurant operations
func (r *restaurantRepository) CreateRestaurant(restaurant *model.Restaurant) error {
	result := r.db.Create(restaurant)
	if result.Error != nil {
		return fmt.Errorf("failed to create restaurant: %v", result.Error)
	}
	return nil
}

func (r *restaurantRepository) GetRestaurantByEmail(email string) (*model.Restaurant, error) {
	var restaurant model.Restaurant
	result := r.db.Where("owner_email = ?", email).First(&restaurant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, model.ErrRestaurantNotFound
		}
		return nil, result.Error
	}
	return &restaurant, nil
}

func (r *restaurantRepository) GetRestaurantByID(id string) (*model.Restaurant, error) {
	var restaurant model.Restaurant
	result := r.db.Where("id = ?", id).First(&restaurant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, model.ErrRestaurantNotFound
		}
		return nil, result.Error
	}
	return &restaurant, nil
}

func (r *restaurantRepository) UpdateRestaurant(restaurant *model.Restaurant) error {
	result := r.db.Save(restaurant)
	if result.Error != nil {
		return fmt.Errorf("failed to update restaurant: %v", result.Error)
	}
	return nil
}

func (r *restaurantRepository) GetAllRestaurants() ([]*model.Restaurant, error) {
	var restaurants []*model.Restaurant
	result := r.db.Find(&restaurants)
	if result.Error != nil {
		return nil, result.Error
	}
	return restaurants, nil
}

func (r *restaurantRepository) BanRestaurant(restaurantID, reason string) error {
	result := r.db.Model(&model.Restaurant{}).
		Where("id = ?", restaurantID).
		Updates(map[string]interface{}{
			"is_banned":  true,
			"ban_reason": reason,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrRestaurantNotFound
	}
	return nil
}

func (r *restaurantRepository) UnbanRestaurant(restaurantID string) error {
	result := r.db.Model(&model.Restaurant{}).
		Where("id = ?", restaurantID).
		Updates(map[string]interface{}{
			"is_banned":  false,
			"ban_reason": "",
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrRestaurantNotFound
	}
	return nil
}

// Product operations
func (r *restaurantRepository) AddProduct(product *model.Product) error {
	result := r.db.Create(product)
	if result.Error != nil {
		return fmt.Errorf("failed to add product: %v", result.Error)
	}
	return nil
}

func (r *restaurantRepository) GetProductByID(productID string) (*model.Product, error) {
	var product model.Product
	result := r.db.Where("id = ?", productID).First(&product)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, model.ErrProductNotFound
		}
		return nil, result.Error
	}
	return &product, nil
}

func (r *restaurantRepository) GetProductsByRestaurantID(restaurantID string) ([]*model.Product, error) {
	var products []*model.Product
	result := r.db.Where("restaurant_id = ?", restaurantID).Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

func (r *restaurantRepository) UpdateProduct(product *model.Product) error {
	result := r.db.Save(product)
	if result.Error != nil {
		return fmt.Errorf("failed to update product: %v", result.Error)
	}
	return nil
}

func (r *restaurantRepository) DeleteProduct(productID string) error {
	result := r.db.Delete(&model.Product{}, "id = ?", productID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrProductNotFound
	}
	return nil
}

func (r *restaurantRepository) GetAllProducts() ([]*model.Product, error) {
	var products []*model.Product
	result := r.db.Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

func (r *restaurantRepository) UpdateProductStock(productID string, quantity int32) error {
	result := r.db.Model(&model.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrProductNotFound
	}
	return nil
}

func (r *restaurantRepository) GetProductStock(productID string) (int32, error) {
	var product model.Product
	result := r.db.Select("stock").Where("id = ?", productID).First(&product)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, model.ErrProductNotFound
		}
		return 0, result.Error
	}
	return product.Stock, nil
}

func (r *restaurantRepository) GetRestaurantWithProducts(restaurantID string) (*model.Restaurant, []*model.Product, error) {
	var restaurant model.Restaurant
	result := r.db.Where("id = ?", restaurantID).First(&restaurant)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil, model.ErrRestaurantNotFound
		}
		return nil, nil, result.Error
	}

	var products []*model.Product
	result = r.db.Where("restaurant_id = ?", restaurantID).Find(&products)
	if result.Error != nil {
		return nil, nil, result.Error
	}

	return &restaurant, products, nil
}

func (r *restaurantRepository) GetAllRestaurantsWithProducts() ([]*model.Restaurant, error) {
	var restaurants []*model.Restaurant
	result := r.db.Find(&restaurants)
	if result.Error != nil {
		return nil, result.Error
	}
	return restaurants, nil
}
