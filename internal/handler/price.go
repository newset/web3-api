package handler

import (
	"net/http"
	"strings"
	"web3-api/internal/model"

	"github.com/gin-gonic/gin"
)

// getNativePrice godoc
// @Summary      查询本币 USD 价格
// @Description  通过 Binance 获取指定链本币的实时 USD 价格
// @Tags         price
// @Accept       json
// @Produce      json
// @Param        chain path string true "链名称" Enums(eth, eth_sepolia, bnb, bnb_testnet, tron, tron_shasta, solana, solana_devnet)
// @Success      200   {object} model.Response{data=price.PriceInfo}
// @Failure      400   {object} model.Response
// @Failure      500   {object} model.Response
// @Router       /api/v1/chain/{chain}/price [get]
func (h *Handler) getNativePrice(c *gin.Context) {
	chain := c.Param("chain")

	result, err := h.priceSvc.GetPrice(c.Request.Context(), chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

// getNativePrices godoc
// @Summary      批量查询本币 USD 价格
// @Description  批量获取多个链本币的实时 USD 价格，不传 chains 则返回全部链
// @Tags         price
// @Accept       json
// @Produce      json
// @Param        chains query string false "逗号分隔的链名称，如 eth,bnb,tron,solana（为空则返回全部）"
// @Success      200    {object} model.Response{data=[]price.PriceInfo}
// @Failure      500    {object} model.Response
// @Router       /api/v1/prices [get]
func (h *Handler) getNativePrices(c *gin.Context) {
	chainsParam := c.Query("chains")
	var chains []string
	if chainsParam == "" {
		chains = h.priceSvc.AllChains()
	} else {
		chains = strings.Split(chainsParam, ",")
	}

	result, err := h.priceSvc.GetPrices(c.Request.Context(), chains)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}
