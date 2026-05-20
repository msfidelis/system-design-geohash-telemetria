package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/msfidelis01/geoip-teste/api/services"
)

type CarsHandler struct {
	service *services.LocationService
}

func NewCarsHandler(service *services.LocationService) *CarsHandler {
	return &CarsHandler{service: service}
}

func (h *CarsHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	car, err := h.service.GetCarByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if car == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "car not found"})
	}

	return c.JSON(car)
}

func (h *CarsHandler) GetByLocation(c *fiber.Ctx) error {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid lat"})
	}

	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid lon"})
	}

	result, err := h.service.GetCarsByRegion(lat, lon)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
