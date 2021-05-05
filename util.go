package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/lithammer/shortuuid/v3"
)

func validateOrderReq(userData UserData, or OrderReq) error {
	// validate userId is present in db
	if userData.userId == "" {
		return fmt.Errorf("userId: %s not an actual user", or.UserId)
	}
	// validate user owns asset
	if _, ok := userData.assets[or.AssetId]; !ok {
		return fmt.Errorf("user doesn't own AssetId:%s", or.AssetId)
	}
	// validate user has enough cash to buy
	if or.BuyOrSell == BUY && userData.cash < getTotalAssetCost(or.Limit, or.Size) {
		return errors.New("user doesn't have enough cash")
	}
	// validate user has enough assets to sell
	if or.BuyOrSell == SELL && userData.assets[or.AssetId] < or.Size {
		return errors.New("user doesn't have enough assets to sell")
	}

	return nil
}

// createOrderId creates a unique order id
// Note, in a prod environment, this could be improved to ensure uniqueness in a distributed system.
func createOrderId() OrderId {
	return OrderId(shortuuid.New())
}

// createOrderFromOrderReq creates an Order{} struct given an orderReq struct{}
func createOrderFromOrderReq(or OrderReq) Order {
	oid := createOrderId()
	return Order{
		orderId:   oid,
		userId:    or.UserId,
		limit:     or.Limit,
		assetId:   or.AssetId,
		size:      or.Size,
		buyOrSell: or.BuyOrSell,
		eventAt:   time.Now(),
		status:    Working,
	}
}

func getTotalAssetCost(limit Usd, size int) Usd {
	return Usd(int(limit) * size)
}

func min(a,b int) int {
	if a < b {
		return a
	}
	return b
}

func orderToOrderResp(order Order) OrderResp {
	return OrderResp{
		OrderId: order.orderId,
		UserId: order.userId,
		Limit: order.limit,
		AssetId: order.assetId,
		Size: order.size,
		BuyOrSell: order.buyOrSell,
		EventAt: order.eventAt,
		Status: order.status,
		Filled: order.filled,
	}
}
