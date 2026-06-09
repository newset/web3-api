package handler

import (
	"net/http"
	"web3-api/internal/model"

	"github.com/gin-gonic/gin"
)

// getNativeTransactions godoc
// @Summary      查询本币交易记录
// @Description  查询指定链上某个地址的本币（ETH/BNB/TRX/SOL）交易记录
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        chain   path   string true  "链名称 (主网: eth, bnb, tron, solana / 测试网: eth_sepolia, bnb_testnet, tron_shasta, solana_devnet)"  Enums(eth, eth_sepolia, bnb, bnb_testnet, tron, tron_shasta, solana, solana_devnet)
// @Param        address query  string true  "钱包地址"
// @Param        cursor  query  string false "分页游标"
// @Success      200     {object} model.Response{data=model.TransactionPage}
// @Failure      400     {object} model.Response
// @Failure      500     {object} model.Response
// @Router       /api/v1/chain/{chain}/transactions [get]
func (h *Handler) getNativeTransactions(c *gin.Context) {
	chain := c.Param("chain")
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "address is required"})
		return
	}
	cursor := c.Query("cursor")

	result, err := h.svc.GetNativeTransactions(c.Request.Context(), chain, address, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

// getTokenTransactions godoc
// @Summary      查询代币交易记录
// @Description  查询指定链上某个地址针对特定代币的转账记录
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        chain    path   string true  "链名称 (主网: eth, bnb, tron, solana / 测试网: eth_sepolia, bnb_testnet, tron_shasta, solana_devnet)"  Enums(eth, eth_sepolia, bnb, bnb_testnet, tron, tron_shasta, solana, solana_devnet)
// @Param        address  query  string true  "钱包地址"
// @Param        contract query  string true  "代币合约地址"
// @Param        cursor   query  string false "分页游标"
// @Success      200      {object} model.Response{data=model.TransactionPage}
// @Failure      400      {object} model.Response
// @Failure      500      {object} model.Response
// @Router       /api/v1/chain/{chain}/token/transactions [get]
func (h *Handler) getTokenTransactions(c *gin.Context) {
	chain := c.Param("chain")
	address := c.Query("address")
	contract := c.Query("contract")
	if address == "" || contract == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "address and contract are required"})
		return
	}
	cursor := c.Query("cursor")

	result, err := h.svc.GetTransactions(c.Request.Context(), chain, address, contract, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}
