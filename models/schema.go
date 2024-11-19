package model

type Restaurant struct {
    ID           string
    OwnerEmail   string
    PasswordHash string
    Name         string
    PhoneNumber  uint64
    IsBanned     bool
    BanReason    string
    Address      *Address
}

type Address struct {
    ID         string
    StreetName string
    Locality   string
    State      string
    Pincode    string
}

type Product struct {
    ID           string
    RestaurantID string
    Name         string
    Description  string
    Price        float64
    Stock        int32
    Category     string
}