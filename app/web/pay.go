package web

import (
	"USDTMore/app/log"
	"USDTMore/app/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"time"
)

/*
生成支付页面
*/
func CheckoutCounter(ctx *gin.Context) {
	var tradeId = ctx.Param("trade_id")
	var order, ok = model.GetTradeOrder(tradeId)
	if !ok {
		ctx.String(200, "订单不存在")
		return
	}

	uri, err := url.ParseRequestURI(order.ReturnUrl)
	if err != nil {
		ctx.String(200, "同步地址错误")
		log.Error("同步地址解析错误", err.Error())
		return
	}

	var coinChain = "TRC20"
	if order.Chain == "POLY" {
		coinChain = "Polygon"
	}
	if order.Chain == "OP" {
		coinChain = "Optimism"
	}
	if order.Chain == "BSC" {
		coinChain = "BEP20"
	}

	vars := gin.H{
		"http_host":  uri.Host,
		"trade_id":   tradeId,
		"amount":     order.Amount,
		"chain":      order.Chain,
		"coin_chain": coinChain,
		"address":    order.Address,
		"expire":     int64(order.ExpiredAt.Sub(time.Now()).Seconds()),
		"return_url": order.ReturnUrl,
		"usdt_rate":  order.UsdtRate,
	}
	ctx.HTML(200, "checkout-counter.html", vars)
}

/*
网页查看订单状态
*/
func CheckStatus(ctx *gin.Context) {
	var tradeId = ctx.Param("trade_id")
	var order, ok = model.GetTradeOrder(tradeId)
	if !ok {
		ctx.JSON(200, RespFailJson(fmt.Errorf("订单不存在")))
		return
	}

	var returnUrl string
	if order.Status == model.OrderStatusSuccess {
		returnUrl = order.ReturnUrl
	}

	// 返回响应数据
	ctx.JSON(200, gin.H{"trade_id": tradeId, "status": order.Status, "return_url": returnUrl})
}
