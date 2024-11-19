package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	config "github.com/liju-github/FoodBuddyMicroserviceRestaurant/configs"
	"github.com/liju-github/FoodBuddyMicroserviceRestaurant/db"
	"github.com/liju-github/FoodBuddyMicroserviceRestaurant/repository"
	"github.com/liju-github/FoodBuddyMicroserviceRestaurant/service"
)

func main() {
	// Load configurations
	config := config.LoadConfig()

	// Initialize database
	db, err := db.Connect(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize repository
	repo := repository.NewRestaurantRepository(db)

	// Initialize service
	svc := service.NewRestaurantService(repo)

	// Initialize gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.RESTAURANTGRPCPORT))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	restaurantPb.RegisterRestaurantServiceServer(grpcServer, svc)

	log.Printf("Restaurant Service starting on port %s", config.RESTAURANTGRPCPORT)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
