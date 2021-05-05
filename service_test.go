package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderMatchingService_InitExchange(t *testing.T) {
	s := newOrderMatchingService()
	defer s.Close()

	req := InitExchangeReq{
		UserId: userId1,
		Assets: []Asset{{assetId1, 10}, {assetId2, 15}},
		Cash:   1000,
	}

	s.InitExchange([]InitExchangeReq{req})

	actual := s.Store.GetUserData(req.UserId)
	assert.Equal(t, userId1, actual.userId)
	assert.Equal(t, Usd(1000), actual.cash)
	assert.Equal(t, 10, actual.assets[assetId1])
	assert.Equal(t, 15, actual.assets[assetId2])
	assert.Empty(t, actual.orders)
}

func TestOrderMatchingService_SaveOrderToStore(t *testing.T) {
	s := newOrderMatchingService()
	defer s.Close()

	setupTestUsers(s)

	orderReq := OrderReq{
		UserId:    userId1,
		Limit:     100,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	order := createOrderFromOrderReq(orderReq)

	s.SaveOrderToStore(order)

	actual := s.Store.GetUserData(userId1).orders[order.orderId]
	assert.Equal(t, order, actual)
}

func TestOrderMatchingService_ProcessOrderReqs(t *testing.T) {
	s := newOrderMatchingService()
	defer s.Close()

	setupTestUsers(s)

	buyOrderReq1 := OrderReq{
		UserId:    userId1,
		Limit:     101,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	buyOrderReq2 := OrderReq{
		UserId:    userId1,
		Limit:     100,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	buyOrderReq3 := OrderReq{
		UserId:    userId1,
		Limit:     100,
		AssetId:   assetId2, // different asset to buy
		Size:      10,
		BuyOrSell: BUY,
	}

	s.OCh <- buyOrderReq1
	s.OCh <- buyOrderReq2
	s.OCh <- buyOrderReq3

	time.Sleep(50 * time.Millisecond) // give time for goroutine to process requests

	// assert state of Store is as expected
	userData1 := s.Store.GetUserData(userId1)

	assert.Equal(t, 3, len(userData1.orders))
	assert.Equal(t, Usd(6990), userData1.cash) // assert available cash is updated
	assert.Equal(t, 100, userData1.assets[assetId1])
	assert.Equal(t, 100, userData1.assets[assetId2])

	// assert state of order book
	assert.Equal(t, 2, len(s.OrderBooks.orderBooks)) // assert there are 2 order books, one each for assertId1, assertId2
	buyOrder1 := s.OrderBooks.GetTopOrder(assetId1, BUY)
	buyOrder2 := s.OrderBooks.getOrderBook(assetId1).BuyList.front.next.val
	assert.Equal(t, Usd(101), buyOrder1.limit)
	assert.Equal(t, 10, buyOrder1.size)
	assert.Equal(t, Usd(100), buyOrder2.limit)
	assert.Equal(t, 10, buyOrder2.size)
	buyOrder3 := s.OrderBooks.GetTopOrder(assetId2, BUY)
	assert.Equal(t, Usd(100), buyOrder3.limit)
	assert.Equal(t, 10, buyOrder3.size)

	// create a sell order, assert it executed on a buy order
	sellOrderReq1 := OrderReq{
		UserId:    userId2,
		Limit:     100,
		AssetId:   assetId1,
		Size:      15,
		BuyOrSell: SELL,
	}

	sellOrderReq2 := OrderReq{
		UserId:    userId2,
		Limit:     102,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: SELL,
	}

	sellOrderReq3 := OrderReq{
		UserId:    userId2,
		Limit:     100,
		AssetId:   assetId2,
		Size:      10,
		BuyOrSell: SELL,
	}

	s.OCh <- sellOrderReq1
	s.OCh <- sellOrderReq2
	s.OCh <- sellOrderReq3

	time.Sleep(50 * time.Millisecond) // give time for goroutine to process requests

	// assert state of Store is as expected
	userData1 = s.Store.GetUserData(userId1)

	assert.Equal(t, Complete, userData1.orders[buyOrder1.orderId].status) // assert buy order 1 was completely executed
	assert.Equal(t, Working, userData1.orders[buyOrder2.orderId].status)  // assert buy order 3 was partially executed and still in working status
	assert.Equal(t, Complete, userData1.orders[buyOrder3.orderId].status) // assert buy order 3 was completely executed
	assert.Equal(t, Usd(6990), userData1.cash)                            // assert available cash is stays same
	assert.Equal(t, 115, userData1.assets[assetId1])                      // bought 15 assets of asset1
	assert.Equal(t, 110, userData1.assets[assetId2])                      // bought 10 assets of asset1

	// user2
	userData2 := s.Store.GetUserData(userId2)
	assert.Equal(t, Usd(12500), userData2.cash)
	assert.Equal(t, 75, userData2.assets[assetId1]) // - 15 assets in flight or sold
	assert.Equal(t, 90, userData2.assets[assetId2])
	assert.Equal(t, 3, len(userData2.orders))
	sellOrder1 := s.OrderBooks.getOrderBook(assetId1).SellList.front.val
	assert.Equal(t, Working, userData2.orders[sellOrder1.orderId].status) // assert buy order 1 was completely executed

	// assert state of order book
	assert.Equal(t, 1, s.OrderBooks.getOrderBook(assetId1).BuyList.getSize()) // assert only 1 buy order left for asset1
	topBuyOrderAsset1 := s.OrderBooks.GetTopOrder(assetId1, BUY)
	assert.Equal(t, Usd(100), topBuyOrderAsset1.limit)
	assert.Equal(t, 5, topBuyOrderAsset1.size)
	assert.Equal(t, 0, s.OrderBooks.getOrderBook(assetId2).BuyList.getSize()) // no buy order left for asset2 order book

	assert.Equal(t, 1, s.OrderBooks.getOrderBook(assetId1).BuyList.getSize()) // assert only 1 sell order left for asset1
	topSellOrderAsset1 := s.OrderBooks.GetTopOrder(assetId1, SELL)
	assert.Equal(t, Usd(102), topSellOrderAsset1.limit)
	assert.Equal(t, 10, topSellOrderAsset1.size)

	// perform a buy order to execute a sell order in the order book
	buyOrderReq4 := OrderReq{
		UserId:    userId1,
		Limit:     103,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	s.OCh <- buyOrderReq4

	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, s.OrderBooks.getOrderBook(assetId1).BuyList.getSize())
	assert.Equal(t, 0, s.OrderBooks.getOrderBook(assetId1).SellList.getSize())
}

func TestOrderMatchingService_GetUserActiveOrders(t *testing.T) {
	s := newOrderMatchingService()
	defer s.Close()

	setupTestUsers(s)

	assert.Empty(t, s.GetUserActiveOrders(userId1)) // user has no orders
	assert.Empty(t, s.GetUserActiveOrders(userId2)) // user not present in db

	buyOrderReq1 := OrderReq{
		UserId:    userId1,
		Limit:     101,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	sellOrderReq2 := OrderReq{
		UserId:    userId1,
		Limit:     120,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: SELL,
	}

	s.OCh <- buyOrderReq1
	s.OCh <- sellOrderReq2

	time.Sleep(5 * time.Millisecond)

	activeOrders := s.GetUserActiveOrders(userId1)
	assert.Equal(t, 2, len(activeOrders))
	for _, order := range activeOrders {
		assert.Equal(t, Working, order.Status)
	}
}

func TestOrderMatchingService_GetUserCompleteOrder(t *testing.T) {
	s := newOrderMatchingService()
	defer s.Close()

	setupTestUsers(s)

	assert.Empty(t, s.GetUserCompleteOrders(userId1)) // user has no orders
	assert.Empty(t, s.GetUserCompleteOrders(userId1)) // user not present in db

	buyOrderReq1 := OrderReq{
		UserId:    userId1,
		Limit:     101,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	buyOrderReq2 := OrderReq{
		UserId:    userId1,
		Limit:     99,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	sellOrderReq2 := OrderReq{
		UserId:    userId2,
		Limit:     100,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: SELL,
	}

	s.OCh <- buyOrderReq1
	s.OCh <- buyOrderReq2
	s.OCh <- sellOrderReq2

	time.Sleep(5 * time.Millisecond)

	completeOrders := s.GetUserCompleteOrders(userId1)
	assert.Equal(t, 1, len(completeOrders))
	assert.Equal(t, Usd(101), completeOrders[0].Limit)
	assert.Equal(t, BUY, completeOrders[0].BuyOrSell)

	completeOrders2 := s.GetUserCompleteOrders(userId2)
	assert.Equal(t, 1, len(completeOrders2))
	assert.Equal(t, Usd(100), completeOrders2[0].Limit)
	assert.Equal(t, SELL, completeOrders2[0].BuyOrSell)
}

func TestOrderMatchingService_CancelUserOrder(t *testing.T) {
	s := newOrderMatchingService()
	defer s.Close()

	setupTestUsers(s)

	assert.Empty(t, s.CancelUserOrder(userId1, "wrong")) // user has no orders
	assert.Empty(t, s.CancelUserOrder(userId2, "wrong")) // user not present in db

	buyOrderReq := OrderReq{
		UserId:    userId1,
		Limit:     101,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: BUY,
	}

	sellOrderReq := OrderReq{
		UserId:    userId1,
		Limit:     120,
		AssetId:   assetId1,
		Size:      10,
		BuyOrSell: SELL,
	}

	s.OCh <- buyOrderReq
	s.OCh <- sellOrderReq

	time.Sleep(5 * time.Millisecond)

	// assert state of order book
	assert.Equal(t, 1, s.OrderBooks.getOrderBook(assetId1).BuyList.getSize())
	assert.Equal(t, 1, s.OrderBooks.getOrderBook(assetId1).SellList.getSize())

	activeOrders := s.GetUserActiveOrders(userId1)
	buyOrder := getOrder(activeOrders, BUY)
	sellOrder := getOrder(activeOrders, SELL)

	s.CancelUserOrder(userId1, buyOrder.OrderId)

	// assert store data
	activeOrders = s.GetUserActiveOrders(userId1)
	assert.Equal(t, 1, len(activeOrders))
	assert.Equal(t, sellOrder.OrderId, activeOrders[0].OrderId)
	assert.Equal(t, Usd(10000), s.Store.GetUserData(buyOrder.UserId).cash)
	assert.Equal(t, Canceled, s.Store.GetUserData(buyOrder.UserId).orders[buyOrder.OrderId].status)

	// assert state of order book
	assert.Equal(t, 0, s.OrderBooks.getOrderBook(assetId1).BuyList.getSize())
	assert.Equal(t, 1, s.OrderBooks.getOrderBook(assetId1).SellList.getSize())
}

func setupTestUsers(s *OrderMatchingService) {
	req1 := InitExchangeReq{
		UserId: userId1,
		Assets: []Asset{{assetId1, 100}, {assetId2, 100}},
		Cash:   10000,
	}

	req2 := InitExchangeReq{
		UserId: userId2,
		Assets: []Asset{{assetId1, 100}, {assetId2, 100}},
		Cash:   10000,
	}

	s.InitExchange([]InitExchangeReq{req1, req2})
}

func getOrder(orders []OrderResp, orderType BuyOrSell) OrderResp {
	for _, o := range orders {
		if o.BuyOrSell == orderType {
			return o
		}
	}
	return OrderResp{}
}
