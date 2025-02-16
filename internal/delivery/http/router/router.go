package router

import (
	"github.com/1way-market/v3/internal/delivery/http/handler"
	"github.com/1way-market/v3/internal/usecase"
	"github.com/gin-gonic/gin"
)

func Setup(useCases *usecase.UseCases) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v3 routes
	v3 := r.Group("/v3")
	{
		adHandler := handler.NewAdHandler(useCases.AdUseCase)
		ads := v3.Group("/ads")
		{
			ads.GET("", adHandler.GetAds)
			ads.POST("", adHandler.CreateAd)
			ads.PUT("/:id", adHandler.UpdateAd)
			ads.DELETE("/:id", adHandler.DeleteAd)
		}
	}

	return r
}
