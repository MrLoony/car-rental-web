package service

import (
	"context"
	"fmt"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
)

type CarService struct {
	repo *repository.CarRepository
}

func NewCarService(repo *repository.CarRepository) *CarService {
	return &CarService{repo: repo}
}

func (s *CarService) ListAvailableCars(ctx context.Context) ([]model.Car, error) {
	cars, err := s.repo.ListAvailableCars(ctx)
	if err != nil {
		return nil, fmt.Errorf("list available cars: %w", err)
	}

	return cars, nil
}

func (s *CarService) GetCarBySlug(ctx context.Context, slug string) (model.Car, error) {
	car, err := s.repo.GetCarBySlug(ctx, slug)
	if err != nil {
		return model.Car{}, fmt.Errorf("get car by slug: %w", err)
	}

	return car, nil
}
