package handler

import (
	"web3-api/internal/price"
	"web3-api/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	svc      *service.Service
	priceSvc *price.Service
}

func New(svc *service.Service, priceSvc *price.Service) *Handler {
	return &Handler{svc: svc, priceSvc: priceSvc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/chain/:chain/tokens", h.getTokens)
		v1.GET("/chain/:chain/transactions", h.getNativeTransactions)
		v1.GET("/chain/:chain/token/transactions", h.getTokenTransactions)
		v1.GET("/chain/:chain/nfts", h.getNFTs)
		v1.GET("/chain/:chain/price", h.getNativePrice)
		v1.GET("/prices", h.getNativePrices)
	}

	// Swagger UI: http://localhost:8080/swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
