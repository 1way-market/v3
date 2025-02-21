package usecase

import (
	"context"
	"fmt"
	"time"

	"encoding/json"
	"github.com/1way-market/v3/internal/domain"
	"github.com/go-redis/redis/v8"
)

type AdRepository interface {
	FindWithFilter(ctx context.Context, filter domain.FilterRequest) (*domain.PaginatedResponse, error)
	Create(ctx context.Context, ad *domain.Ad) error
	Update(ctx context.Context, ad *domain.Ad) error
	Delete(ctx context.Context, id uint) error
}

type AdUseCase struct {
	repo  AdRepository
	cache *redis.Client
}

func NewAdUseCase(repo AdRepository, cache *redis.Client) *AdUseCase {
	return &AdUseCase{
		repo:  repo,
		cache: cache,
	}
}

func (uc *AdUseCase) GetAds(ctx context.Context, filter domain.FilterRequest) (*domain.PaginatedResponse, error) {
	// Try to get from cache first
	cacheKey := uc.buildCacheKey(filter)
	if cachedData, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		var response domain.PaginatedResponse
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			return &response, nil
		}
	}

	// Get from database
	response, err := uc.repo.FindWithFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if jsonData, err := json.Marshal(response); err == nil {
		uc.cache.Set(ctx, cacheKey, jsonData, 5*time.Minute)
	}

	return response, nil
}

func (uc *AdUseCase) buildCacheKey(filter domain.FilterRequest) string {
	key := fmt.Sprintf("ads:filter:%v:%v:%v:%v:%v",
		filter.CategoryIDs,
		filter.TextSearch,
		filter.SortBy,
		filter.PageToken,
		filter.PageSize,
	)

	for _, prop := range filter.PropertyFilters {
		key += fmt.Sprintf(":%v=%v", prop.PropertyID, prop.Values)
	}

	return key
}

func (uc *AdUseCase) CreateAd(ctx context.Context, ad *domain.Ad) error {
	if err := uc.repo.Create(ctx, ad); err != nil {
		return err
	}

	// Invalidate relevant cache entries
	uc.cache.Del(ctx, "ads:*")
	return nil
}

func (uc *AdUseCase) UpdateAd(ctx context.Context, ad *domain.Ad) error {
	if err := uc.repo.Update(ctx, ad); err != nil {
		return err
	}

	// Invalidate relevant cache entries
	uc.cache.Del(ctx, "ads:*")
	return nil
}

func (uc *AdUseCase) DeleteAd(ctx context.Context, id uint) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate relevant cache entries
	uc.cache.Del(ctx, "ads:*")
	return nil
}
