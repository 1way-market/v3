package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/1way-market/v3/internal/domain"
	"github.com/gin-gonic/gin"
)

type AdUseCase interface {
	GetAds(ctx context.Context, filter domain.FilterRequest) (*domain.PaginatedResponse, error)
	CreateAd(ctx context.Context, ad *domain.Ad) error
	UpdateAd(ctx context.Context, ad *domain.Ad) error
	DeleteAd(ctx context.Context, id uint) error
}

type AdHandler struct {
	useCase AdUseCase
}

func NewAdHandler(useCase AdUseCase) *AdHandler {
	return &AdHandler{useCase: useCase}
}

// @Summary Get filtered ads
// @Description Get a paginated list of ads with filters
// @Tags ads
// @Accept json
// @Produce json
// @Param categories query []int false "Category IDs"
// @Param properties query object false "Dynamic properties filter"
// @Param q query string false "Text search"
// @Param sort query string false "Sort order (price_asc, price_desc, date_desc)"
// @Param next_page query string false "Page token for pagination"
// @Param page_size query int false "Number of items per page"
// @Param lang query string true "Language code (e.g., 'ru', 'en')"
// @Success 200 {object} domain.PaginatedResponse
// @Router /v3/ads [get]
func (h *AdHandler) GetAds(c *gin.Context) {
	var filter domain.FilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.useCase.GetAds(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Create new ad
// @Description Create a new advertisement
// @Tags ads
// @Accept json
// @Produce json
// @Param ad body domain.Ad true "Advertisement object"
// @Success 201 {object} domain.Ad
// @Router /v3/ads [post]
func (h *AdHandler) CreateAd(c *gin.Context) {
	var ad domain.Ad
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.CreateAd(c.Request.Context(), &ad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ad)
}

// @Summary Update ad
// @Description Update an existing advertisement
// @Tags ads
// @Accept json
// @Produce json
// @Param id path int true "Advertisement ID"
// @Param ad body domain.Ad true "Advertisement object"
// @Success 200 {object} domain.Ad
// @Router /v3/ads/{id} [put]
func (h *AdHandler) UpdateAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var ad domain.Ad
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ad.ID = uint(id)
	if err := h.useCase.UpdateAd(c.Request.Context(), &ad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ad)
}

// @Summary Delete ad
// @Description Delete an advertisement
// @Tags ads
// @Produce json
// @Param id path int true "Advertisement ID"
// @Success 204 "No Content"
// @Router /v3/ads/{id} [delete]
func (h *AdHandler) DeleteAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.useCase.DeleteAd(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
