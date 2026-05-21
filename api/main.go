package main

import (
	"fmt"
	"os"

	"app/handlers"
	cassandrapkg "app/pkg/cassandra"
	"app/services"

	"github.com/gofiber/fiber/v2"
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
