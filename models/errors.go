package model

import "errors"

var (
    ErrRestaurantNotFound     = errors.New("restaurant not found")
    ErrProductNotFound        = errors.New("product not found")
    ErrInvalidCredentials     = errors.New("invalid credentials")
    ErrEmailAlreadyExists     = errors.New("email already exists")
    ErrRestaurantIsBanned     = errors.New("restaurant is banned")
    ErrInsufficientStock      = errors.New("insufficient stock")
    ErrInvalidStockOperation  = errors.New("invalid stock operation")
)