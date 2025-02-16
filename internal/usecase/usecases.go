package usecase

import (
	"github.com/1way-market/v3/internal/repository"
	"github.com/go-redis/redis/v8"
)

type UseCases struct {
	AdUseCase *AdUseCase
}

func NewUseCases(repos *repository.Repositories, redisClient *redis.Client) *UseCases {
	return &UseCases{
		AdUseCase: NewAdUseCase(repos.Ad, redisClient),
	}
}
