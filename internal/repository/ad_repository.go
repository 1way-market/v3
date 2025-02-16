package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/1way-market/v3/internal/domain"
	"gorm.io/gorm"
)

type AdRepository struct {
	db *gorm.DB
}

func NewAdRepository(db *gorm.DB) *AdRepository {
	return &AdRepository{db: db}
}

func (r *AdRepository) FindWithFilter(ctx context.Context, filter domain.FilterRequest) (*domain.PaginatedResponse, error) {
	var ads []domain.Ad
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&domain.Ad{})

	// Apply category filter
	if len(filter.CategoryIDs) > 0 {
		query = query.Where("category_ids && ?", filter.CategoryIDs)
	}

	// Apply text search if provided
	if filter.TextSearch != "" {
		query = query.Where("search_vector @@ plainto_tsquery(?)", filter.TextSearch)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Apply dynamic property filters
	for prop, values := range filter.Properties {
		if len(values) > 0 {
			query = query.Where(fmt.Sprintf("attributes->>'%s' IN (?)", prop), values)
		}
	}

	// Apply price filters
	if filter.MinPrice != nil || filter.MaxPrice != nil || filter.Currency != "" {
		if filter.Currency != "" {
			query = query.Where("price->>'currency' = ?", filter.Currency)
		}
		if filter.MinPrice != nil {
			query = query.Where("(price->>'value')::float >= ?", *filter.MinPrice)
		}
		if filter.MaxPrice != nil {
			query = query.Where("(price->>'value')::float <= ?", *filter.MaxPrice)
		}
	}

	// Apply sorting
	switch filter.SortBy {
	case "price_asc":
		query = query.Order("(price->>'value')::float ASC NULLS LAST")
	case "price_desc":
		query = query.Order("(price->>'value')::float DESC NULLS LAST")
	case "date_desc":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Count total results
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	pageSize := filter.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	if filter.PageToken != "" {
		var lastAd domain.Ad
		if err := r.db.First(&lastAd, "id = ?", filter.PageToken).Error; err != nil {
			return nil, err
		}
		query = query.Where("id > ?", lastAd.ID)
	}

	// Execute query
	if err := query.Limit(pageSize + 1).Find(&ads).Error; err != nil {
		return nil, err
	}

	// Prepare response
	response := &domain.PaginatedResponse{
		TotalCount: totalCount,
	}

	if len(ads) > pageSize {
		response.Items = ads[:pageSize]
		response.NextPage = fmt.Sprintf("%d", ads[pageSize-1].ID)
	} else {
		response.Items = ads
	}

	return response, nil
}

func (r *AdRepository) buildSearchVector(ad *domain.Ad) string {
	// Build search vector from all language versions
	var searchTexts []string

	// Add title texts
	for _, t := range ad.Title {
		searchTexts = append(searchTexts, t.Text)
	}

	// Add description texts if present
	for _, d := range ad.Description {
		searchTexts = append(searchTexts, d.Text)
	}

	// Join all texts with spaces and convert to tsvector
	return fmt.Sprintf("to_tsvector('simple', %s)",
		r.db.Dialector.Explain("?", strings.Join(searchTexts, " ")))
}

func (r *AdRepository) Create(ctx context.Context, ad *domain.Ad) error {
	// Set search vector
	searchVector := r.buildSearchVector(ad)

	// Create ad with all fields
	result := r.db.WithContext(ctx).Omit("created_at", "updated_at").Create(map[string]interface{}{
		"title":         ad.Title,
		"description":   ad.Description,
		"attributes":    ad.Attributes,
		"category_ids":  ad.CategoryIDs,
		"status":        ad.Status,
		"price":         ad.Price,
		"search_vector": searchVector,
	})

	if result.Error != nil {
		return fmt.Errorf("error creating ad: %v", result.Error)
	}

	return nil
}

func (r *AdRepository) Update(ctx context.Context, ad *domain.Ad) error {
	// Set search vector
	searchVector := r.buildSearchVector(ad)

	result := r.db.WithContext(ctx).Model(&domain.Ad{}).
		Where("id = ?", ad.ID).
		Omit("created_at").
		Updates(map[string]interface{}{
			"title":         ad.Title,
			"description":   ad.Description,
			"attributes":    ad.Attributes,
			"category_ids":  ad.CategoryIDs,
			"status":        ad.Status,
			"price":         ad.Price,
			"search_vector": searchVector,
		})

	if result.Error != nil {
		return fmt.Errorf("error updating ad: %v", result.Error)
	}

	return nil
}

func (r *AdRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Ad{}, id).Error
}

func (r *AdRepository) GetByID(ctx context.Context, id uint) (*domain.Ad, error) {
	var ad domain.Ad
	if err := r.db.WithContext(ctx).First(&ad, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting ad: %v", err)
	}
	return &ad, nil
}

func (r *AdRepository) List(ctx context.Context, filter *domain.FilterRequest) (*domain.PaginatedResponse, error) {
	query := r.db.WithContext(ctx).Model(&domain.Ad{})

	// Apply filters
	if len(filter.CategoryIDs) > 0 {
		query = query.Where("category_ids && ?", filter.CategoryIDs)
	}

	if filter.TextSearch != "" {
		query = query.Where("search_vector @@ plainto_tsquery(?)", filter.TextSearch)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Apply price filters
	if filter.MinPrice != nil || filter.MaxPrice != nil || filter.Currency != "" {
		if filter.Currency != "" {
			query = query.Where("price->>'currency' = ?", filter.Currency)
		}
		if filter.MinPrice != nil {
			query = query.Where("(price->>'value')::float >= ?", *filter.MinPrice)
		}
		if filter.MaxPrice != nil {
			query = query.Where("(price->>'value')::float <= ?", *filter.MaxPrice)
		}
	}

	// Apply sorting
	switch filter.SortBy {
	case "price_asc":
		query = query.Order("(price->>'value')::float ASC NULLS LAST")
	case "price_desc":
		query = query.Order("(price->>'value')::float DESC NULLS LAST")
	case "date_desc":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("error counting ads: %v", err)
	}

	// Apply pagination
	if filter.PageSize > 0 {
		query = query.Limit(filter.PageSize)
	}
	if filter.PageToken != "" {
		// Implement cursor-based pagination using PageToken
		// This is a placeholder - implement according to your needs
	}

	// Get results
	var ads []domain.Ad
	if err := query.Find(&ads).Error; err != nil {
		return nil, fmt.Errorf("error listing ads: %v", err)
	}

	return &domain.PaginatedResponse{
		Items:      ads,
		TotalCount: totalCount,
		// Set NextPage based on your pagination implementation
	}, nil
}
