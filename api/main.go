package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/msfidelis01/geoip-teste/api/handlers"
	cassandrapkg "github.com/msfidelis01/geoip-teste/api/pkg/cassandra"
	"github.com/msfidelis01/geoip-teste/api/services"
)

func cassandraHost() string {
	if v := os.Getenv("CASSANDRA_HOST"); v != "" {
		return v
	}
	return "localhost"
}

func port() string {
	if v := os.Getenv("PORT"); v != "" {
		return v
	}
	return "8080"
}

func main() {
	session := cassandrapkg.Connect(cassandraHost())
	defer session.Close()
	fmt.Printf("connected to cassandra %s\n", cassandraHost())

	locationService := services.NewLocationService(session)
	carsHandler := handlers.NewCarsHandler(locationService)

	app := fiber.New()

	app.Get("/carros/:id", carsHandler.GetByID)
	app.Get("/carros", carsHandler.GetByLocation)

	fmt.Printf("api listening on :%s\n", port())
	if err := app.Listen(":" + port()); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
