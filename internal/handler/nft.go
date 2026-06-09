package handler

import (
	"net/http"
	"web3-api/internal/model"

	"github.com/gin-gonic/gin"
)

// getNFTs godoc
// @Summary      查询地址拥有的 NFT
// @Description  查询指定链上某个地址拥有的所有 NFT
// @Tags         nfts
// @Accept       json
// @Produce      json
// @Param        chain   path   string true  "链名称 (主网: eth, bnb, tron, solana / 测试网: eth_sepolia, bnb_testnet, tron_shasta, solana_devnet)"  Enums(eth, eth_sepolia, bnb, bnb_testnet, tron, tron_shasta, solana, solana_devnet)
// @Param        address query  string true  "钱包地址"
// @Param        cursor  query  string false "分页游标"
// @Success      200     {object} model.Response{data=model.NFTPage}
// @Failure      400     {object} model.Response
// @Failure      500     {object} model.Response
// @Router       /api/v1/chain/{chain}/nfts [get]
func (h *Handler) getNFTs(c *gin.Context) {
	chain := c.Param("chain")
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "address is required"})
		return
	}
	cursor := c.Query("cursor")

	result, err := h.svc.GetNFTs(c.Request.Context(), chain, address, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}
