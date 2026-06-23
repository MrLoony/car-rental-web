package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/service"
)

func TestCarsShowLoadsGalleryImages(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	carRepo := &fakeHandlerCarRepository{
		getBySlugCar: model.Car{
			ID:           42,
			CategoryName: "SUV",
			Brand:        "Nissan",
			Model:        "Patrol",
			Slug:         "nissan-patrol",
			Year:         2024,
			PricePerDay:  140,
			Transmission: "Automatic",
			FuelType:     "Gasoline",
			Seats:        7,
			IsAvailable:  true,
		},
		carImages: []model.CarImage{
			{ID: 1, CarID: 42, ImageURL: "/static/uploads/cars/patrol-front.webp", IsPrimary: true},
		},
	}
	handler.carService = service.NewCarService(carRepo)

	request := requestWithParam(http.MethodGet, "/cars/nissan-patrol", "slug", "nissan-patrol")
	response := httptest.NewRecorder()

	handler.CarsShow().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if carRepo.getCarImagesID != 42 {
		t.Fatalf("gallery car id = %d, want 42", carRepo.getCarImagesID)
	}

	body := response.Body.String()
	if !strings.Contains(body, `data-car-gallery`) {
		t.Fatalf("body does not contain gallery markup:\n%s", body)
	}
	if !strings.Contains(body, `/static/uploads/cars/patrol-front.webp`) {
		t.Fatalf("body does not contain gallery image URL:\n%s", body)
	}
}
