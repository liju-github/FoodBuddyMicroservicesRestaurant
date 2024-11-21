package model

type Restaurant struct {
	ID           string `gorm:"column:id" json:"id"`
	OwnerEmail   string `gorm:"column:owner_email" json:"ownerEmail"`
	PasswordHash string `gorm:"column:password_hash" json:"passwordHash"`
	Name         string `gorm:"column:name" json:"name"`
	PhoneNumber  uint64 `gorm:"column:phone_number" json:"phoneNumber"`
	IsBanned     bool   `gorm:"column:is_banned" json:"isBanned"`
	BanReason    string `gorm:"column:ban_reason" json:"banReason"`
	StreetName   string `gorm:"column:street_name" json:"streetName"`
	Locality     string `gorm:"column:locality" json:"locality"`
	State        string `gorm:"column:state" json:"state"`
	Pincode      string `gorm:"column:pincode" json:"pincode"`
}

type Product struct {
	ID           string  `gorm:"column:id" json:"id"`
	RestaurantID string  `gorm:"column:restaurant_id" json:"restaurantId"`
	Name         string  `gorm:"column:name" json:"name"`
	Description  string  `gorm:"column:description" json:"description"`
	Price        float64 `gorm:"column:price" json:"price"`
	Stock        int32   `gorm:"column:stock" json:"stock"`
	Category     string  `gorm:"column:category" json:"category"`
}
