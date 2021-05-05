package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var userId1 = UserId("userId1")
var userId2 = UserId("userId2")
var assetId1 = AssetId("COIN")
var assetId2 = AssetId("GAME")

func TestOrderBooks_AddOrder(t *testing.T) {
	o1 := Order{orderId: "1", assetId: "COIN", limit: 100, buyOrSell: BUY}

	ob := newOrderBooks()
	ob.AddOrder(o1)

	actual := ob.orderBooks[o1.assetId].BuyList.getOrder(o1.orderId)

	assert.Equal(t, o1, actual)
}

func TestOrderBooks_GetTopOrder(t *testing.T) {
	o1 := Order{orderId: "1", assetId: "COIN", limit: 100, buyOrSell: BUY}
	o2 := Order{orderId: "2", assetId: "COIN", limit: 200, buyOrSell: BUY}

	ob := newOrderBooks()
	ob.AddOrder(o1)
	ob.AddOrder(o2)

	actual := ob.GetTopOrder("COIN", BUY)

	assert.Equal(t, o2, actual)
}

func TestOrderBooks_UpdateOrder(t *testing.T) {
	o1 := Order{orderId: "1", assetId: "COIN", limit: 100, size: 30, buyOrSell: BUY}

	ob := newOrderBooks()
	ob.AddOrder(o1)

	o1.size = 20
	ob.UpdateOrder(o1)

	actual := ob.getOrderBook("COIN").BuyList.getOrder(o1.orderId)

	assert.Equal(t, 20, actual.size)
}

func TestOrderBooks_DeleteOrder(t *testing.T) {
	o1 := Order{orderId: "1", assetId: "COIN", limit: 100, buyOrSell: BUY}

	ob := newOrderBooks()
	ob.AddOrder(o1)

	ob.DeleteOrder(o1)

	actual := ob.getOrderBook("COIN").BuyList.getOrder(o1.orderId)
	lengthOfBuyList := ob.getOrderBook("COIN").BuyList.getSize()

	assert.Empty(t, actual)
	assert.Equal(t, 0, lengthOfBuyList)
}

// Execute top order once
func TestOrderBooks_ExecuteOrder_BUY_CASE1(t *testing.T) {
	sellOrder1 := Order{orderId: "so1", userId: userId1, assetId: assetId1, limit: 100, size: 10, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	sellOrder2 := Order{orderId: "so2", userId: userId1, assetId: assetId1, limit: 101, size: 30, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	buyOrder1 := Order{orderId: "bo1", userId: userId2, assetId: assetId1, limit: 100, size: 25, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	store := setupTestData([]Order{sellOrder1, sellOrder2}, []Order{buyOrder1})

	ob := newOrderBooks()

	ob.AddOrder(sellOrder1)
	ob.AddOrder(sellOrder2)

	ob.ExecuteOrder(buyOrder1, store)

	seller := store.GetUserData(userId1)
	buyer := store.GetUserData(userId2)

	assert.Equal(t, Usd(11000), seller.cash)     // assert user with sell order has increase in cash available
	assert.Equal(t, 110, buyer.assets[assetId1]) // assert user with buy order has increase in asset size available
	assert.Equal(t, Complete, seller.orders[sellOrder1.orderId].status)
	assert.Equal(t, Working, seller.orders[sellOrder2.orderId].status)
	assert.Equal(t, Working, buyer.orders[buyOrder1.orderId].status)

	orderBook := ob.getOrderBook(assetId1)

	assert.Equal(t, 1, orderBook.SellList.getSize()) // assert sellorder2 is in order book
	assert.Equal(t, sellOrder2.orderId, orderBook.SellList.GetTopOrder().orderId)
	assert.Equal(t, 1, orderBook.BuyList.getSize()) // assert partial buy order is in order book
	assert.Equal(t, buyOrder1.orderId, orderBook.BuyList.GetTopOrder().orderId)
	assert.Equal(t, 15, orderBook.BuyList.GetTopOrder().size) // assert remaining buy order size
}

// Execute everything in the sell order list and buy order fully executed
func TestOrderBooks_ExecuteOrder_BUY_CASE2(t *testing.T) {
	sellOrder1 := Order{orderId: "so1", userId: userId1, assetId: assetId1, limit: 100, size: 10, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	sellOrder2 := Order{orderId: "so2", userId: userId1, assetId: assetId1, limit: 100, size: 30, buyOrSell: SELL, eventAt: time.Now().Add(1 * time.Minute), status: Working}
	buyOrder1 := Order{orderId: "bo1", userId: userId2, assetId: assetId1, limit: 100, size: 40, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	store := setupTestData([]Order{sellOrder1, sellOrder2}, []Order{buyOrder1})

	ob := newOrderBooks()

	ob.AddOrder(sellOrder1)
	ob.AddOrder(sellOrder2)

	ob.ExecuteOrder(buyOrder1, store)

	seller := store.GetUserData(userId1)
	buyer := store.GetUserData(userId2)

	assert.Equal(t, Usd(14000), seller.cash)     // assert user with sell order has increase in cash available
	assert.Equal(t, 140, buyer.assets[assetId1]) // assert user with buy order has increase in asset size available
	assert.Equal(t, Complete, seller.orders[sellOrder1.orderId].status)
	assert.Equal(t, Complete, seller.orders[sellOrder2.orderId].status)
	assert.Equal(t, Complete, buyer.orders[buyOrder1.orderId].status)

	orderBook := ob.getOrderBook(assetId1)

	assert.Equal(t, 0, orderBook.SellList.getSize()) // assert sell list is empty
	assert.Equal(t, 0, orderBook.BuyList.getSize())  // assert buy list is empty
}

// Execute all sell orders and add remaining buy order to order book
func TestOrderBooks_ExecuteOrder_BUY_CASE3(t *testing.T) {
	sellOrder1 := Order{orderId: "so1", userId: userId1, assetId: assetId1, limit: 100, size: 10, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	sellOrder2 := Order{orderId: "so2", userId: userId1, assetId: assetId1, limit: 100, size: 30, buyOrSell: SELL, eventAt: time.Now().Add(1 * time.Minute), status: Working}
	buyOrder1 := Order{orderId: "bo1", userId: userId2, assetId: assetId1, limit: 100, size: 50, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	store := setupTestData([]Order{sellOrder1, sellOrder2}, []Order{buyOrder1})

	ob := newOrderBooks()

	ob.AddOrder(sellOrder1)
	ob.AddOrder(sellOrder2)

	ob.ExecuteOrder(buyOrder1, store)

	seller := store.GetUserData(userId1)
	buyer := store.GetUserData(userId2)

	assert.Equal(t, Usd(14000), seller.cash)
	assert.Equal(t, 140, buyer.assets[assetId1])
	assert.Equal(t, Complete, seller.orders[sellOrder1.orderId].status)
	assert.Equal(t, Complete, seller.orders[sellOrder2.orderId].status)
	assert.Equal(t, Working, buyer.orders[buyOrder1.orderId].status)
	assert.Equal(t, 40, buyer.orders[buyOrder1.orderId].filled) // assert amount of assets filled

	orderBook := ob.getOrderBook(assetId1)

	assert.Equal(t, 0, orderBook.SellList.getSize()) // assert sellorder2 is in order book
	assert.Equal(t, 1, orderBook.BuyList.getSize())  // assert partial buy order is in order book
	assert.Equal(t, buyOrder1.orderId, orderBook.BuyList.GetTopOrder().orderId)
	assert.Equal(t, 10, orderBook.BuyList.GetTopOrder().size) // assert remaining buy order size
}

func TestOrderBooks_ExecuteOrder_SELL_CASE1(t *testing.T) {
	buyOrder1 := Order{orderId: "bo1", userId: userId1, assetId: assetId1, limit: 100, size: 10, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	buyOrder2 := Order{orderId: "bo2", userId: userId1, assetId: assetId1, limit: 101, size: 30, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	sellOrder1 := Order{orderId: "so1", userId: userId2, assetId: assetId1, limit: 100, size: 35, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	store := setupTestData([]Order{buyOrder1, buyOrder2}, []Order{sellOrder1})

	ob := newOrderBooks()

	ob.AddOrder(buyOrder1)
	ob.AddOrder(buyOrder2)

	ob.ExecuteOrder(sellOrder1, store)

	buyer := store.GetUserData(userId1)
	seller := store.GetUserData(userId2)

	assert.Equal(t, 135, buyer.assets[assetId1]) // assert buyer's asset's size has increased
	assert.Equal(t, Usd(13500), seller.cash)     // assert seller has increase in cash available
	assert.Equal(t, Complete, buyer.orders[buyOrder2.orderId].status)
	assert.Equal(t, Working, buyer.orders[buyOrder1.orderId].status)
	assert.Equal(t, Complete, seller.orders[sellOrder1.orderId].status)
	assert.Equal(t, 35, seller.orders[sellOrder1.orderId].filled)
	assert.Equal(t, 5, buyer.orders[buyOrder1.orderId].filled)
	assert.Equal(t, 30, buyer.orders[buyOrder2.orderId].filled)

	orderBook := ob.getOrderBook(assetId1)

	assert.Equal(t, 0, orderBook.SellList.getSize())
	assert.Equal(t, 1, orderBook.BuyList.getSize())
	assert.Equal(t, buyOrder1.orderId, orderBook.BuyList.GetTopOrder().orderId)
	assert.Equal(t, 5, orderBook.BuyList.GetTopOrder().size)
}

func TestOrderBooks_ExecuteOrder_SELL_CASE2(t *testing.T) {
	buyOrder1 := Order{orderId: "bo1", userId: userId1, assetId: assetId1, limit: 100, size: 10, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	buyOrder2 := Order{orderId: "bo2", userId: userId1, assetId: assetId1, limit: 101, size: 30, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	sellOrder1 := Order{orderId: "so1", userId: userId2, assetId: assetId1, limit: 100, size: 40, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	store := setupTestData([]Order{buyOrder1, buyOrder2}, []Order{sellOrder1})

	ob := newOrderBooks()

	ob.AddOrder(buyOrder1)
	ob.AddOrder(buyOrder2)

	ob.ExecuteOrder(sellOrder1, store)

	buyer := store.GetUserData(userId1)
	seller := store.GetUserData(userId2)

	assert.Equal(t, 140, buyer.assets[assetId1]) // assert buyer's asset's size has increased
	assert.Equal(t, Usd(14000), seller.cash)     // assert seller has increase in cash available
	assert.Equal(t, Complete, buyer.orders[buyOrder2.orderId].status)
	assert.Equal(t, Complete, buyer.orders[buyOrder1.orderId].status)
	assert.Equal(t, Complete, seller.orders[sellOrder1.orderId].status)

	orderBook := ob.getOrderBook(assetId1)

	assert.Equal(t, 0, orderBook.SellList.getSize())
	assert.Equal(t, 0, orderBook.BuyList.getSize())
}

func TestOrderBooks_ExecuteOrder_SELL_CASE3(t *testing.T) {
	buyOrder1 := Order{orderId: "bo1", userId: userId1, assetId: assetId1, limit: 100, size: 10, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	buyOrder2 := Order{orderId: "bo2", userId: userId1, assetId: assetId1, limit: 101, size: 30, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	sellOrder1 := Order{orderId: "so1", userId: userId2, assetId: assetId1, limit: 100, size: 50, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	store := setupTestData([]Order{buyOrder1, buyOrder2}, []Order{sellOrder1})

	ob := newOrderBooks()

	ob.AddOrder(buyOrder1)
	ob.AddOrder(buyOrder2)

	ob.ExecuteOrder(sellOrder1, store)

	buyer := store.GetUserData(userId1)
	seller := store.GetUserData(userId2)

	assert.Equal(t, 140, buyer.assets[assetId1]) // assert buyer's asset's size has increased
	assert.Equal(t, Usd(14000), seller.cash)     // assert seller has increase in cash available
	assert.Equal(t, Complete, buyer.orders[buyOrder2.orderId].status)
	assert.Equal(t, Complete, buyer.orders[buyOrder1.orderId].status)
	assert.Equal(t, Working, seller.orders[sellOrder1.orderId].status)

	orderBook := ob.getOrderBook(assetId1)

	assert.Equal(t, 1, orderBook.SellList.getSize())
	assert.Equal(t, 0, orderBook.BuyList.getSize())
	assert.Equal(t, sellOrder1.orderId, orderBook.SellList.GetTopOrder().orderId)
	assert.Equal(t, 10, orderBook.SellList.GetTopOrder().size)
}

func TestOrderBooks_ExecuteOrder_SELL_CASE4(t *testing.T) {
	buyOrder1 := Order{orderId: "bo1", userId: userId1, assetId: assetId1, limit: 100, size: 10, buyOrSell: BUY, eventAt: time.Now(), status: Working}
	sellOrder1 := Order{orderId: "so1", userId: userId2, assetId: assetId1, limit: 101, size: 35, buyOrSell: SELL, eventAt: time.Now(), status: Working}
	store := setupTestData([]Order{buyOrder1}, []Order{sellOrder1})

	ob := newOrderBooks()
	ob.AddOrder(buyOrder1)
	ob.ExecuteOrder(sellOrder1, store)

	orderBook := ob.getOrderBook(assetId1)

	assert.Equal(t, 1, orderBook.SellList.getSize())
	assert.Equal(t, 1, orderBook.BuyList.getSize())
}

func TestOrderBook_OrderMatchAvailable_CASE1(t *testing.T) {
	buyOrderList := newOrdersList()
	buyOrder := Order{limit: 1000, buyOrSell: BUY}
	buyOrderList.AddOrder(buyOrder)

	sellOrder1 := Order{limit: 1000, buyOrSell: SELL}
	sellOrder2 := Order{limit: 999, buyOrSell: SELL}
	sellOrder3 := Order{limit: 1001, buyOrSell: SELL}

	assert.True(t, orderMatchAvailable(buyOrderList, sellOrder1, SELL))
	assert.True(t, orderMatchAvailable(buyOrderList, sellOrder2, SELL))
	assert.False(t, orderMatchAvailable(buyOrderList, sellOrder3, SELL))
}

func TestOrderBook_OrderMatchAvailable_CASE2(t *testing.T) {
	sellOrderList := newOrdersList()
	sellOrder := Order{limit: 1000, buyOrSell: BUY}
	sellOrderList.AddOrder(sellOrder)

	buyOrder1 := Order{limit: 1000, buyOrSell: SELL}
	buyOrder2 := Order{limit: 999, buyOrSell: SELL}
	buyOrder3 := Order{limit: 1001, buyOrSell: SELL}

	assert.True(t, orderMatchAvailable(sellOrderList, buyOrder1, BUY))
	assert.False(t, orderMatchAvailable(sellOrderList, buyOrder2, BUY))
	assert.True(t, orderMatchAvailable(sellOrderList, buyOrder3, BUY))
}

func setupTestData(orders1, orders2 []Order) *Store {
	userData1 := UserData{
		userId: userId1,
		cash:   10000,
		assets: map[AssetId]int{assetId1: 100},
		orders: make(map[OrderId]Order),
	}
	userData2 := UserData{
		userId: userId2,
		cash:   10000,
		assets: map[AssetId]int{assetId1: 100},
		orders: make(map[OrderId]Order),
	}

	for _, o := range orders1 {
		userData1.orders[o.orderId] = o
	}
	for _, o := range orders2 {
		userData2.orders[o.orderId] = o
	}

	return &Store{
		db: map[UserId]UserData{userId1: userData1, userId2: userData2},
	}
}
