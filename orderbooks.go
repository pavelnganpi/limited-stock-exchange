package main

import (
	"sync"
)

// OrderBook struct represents an order book for buy and sell orders sorted by price-time priority
type OrderBook struct {
	BuyList  *OrdersList
	SellList *OrdersList
	sync.Mutex // synchronize operations
}

// OrderBooks struct manages all order books for each asset and operations on each asset's order book
type OrderBooks struct {
	orderBooks map[AssetId]*OrderBook
}

func newOrderBooks() *OrderBooks {
	return &OrderBooks{
		orderBooks: make(map[AssetId]*OrderBook),
	}
}

// getOrderBook retrieves the order book for the given assetId
// It creates an empty order book if there isn't one for the given assetId
func (ob *OrderBooks) getOrderBook(assetId AssetId) *OrderBook {
	if _, ok := ob.orderBooks[assetId]; ok {
		return ob.orderBooks[assetId]
	}

	// create new order book for assetId if one isn't present
	ob.orderBooks[assetId] = &OrderBook{
		BuyList:  newOrdersList(),
		SellList: newOrdersList(),
	}
	return ob.orderBooks[assetId]
}

// GetTopOrder returns the top order in the order book for an assetId
func (ob *OrderBooks) GetTopOrder(assetId AssetId, buyOrSell BuyOrSell) Order {
	orderBook := ob.getOrderBook(assetId)
	if orderBook != nil {
		if buyOrSell == BUY {
			return orderBook.BuyList.GetTopOrder()
		} else {
			return orderBook.SellList.GetTopOrder()
		}
	}
	return Order{}
}

// AddOrder adds a new order to the order book for an asset
func (ob *OrderBooks) AddOrder(order Order) {
	orderBook := ob.getOrderBook(order.assetId)
	orderBook.Lock()
	defer orderBook.Unlock()

	if order.buyOrSell == BUY {
		orderBook.BuyList.AddOrder(order)
	} else {
		orderBook.SellList.AddOrder(order)
	}
}

// UpdateOrder updates an order from the order book
func (ob *OrderBooks) UpdateOrder(order Order) {
	orderBook := ob.getOrderBook(order.assetId)
	orderBook.Lock()
	defer orderBook.Unlock()

	if order.buyOrSell == BUY {
		orderBook.BuyList.UpdateOrder(order)
	} else {
		orderBook.SellList.UpdateOrder(order)
	}
}

// DeleteOrder deletes an order from the order book/
func (ob *OrderBooks) DeleteOrder(order Order) {
	orderBook := ob.getOrderBook(order.assetId)
	orderBook.Lock()
	defer orderBook.Unlock()

	if order.buyOrSell == BUY {
		orderBook.BuyList.DeleteOrder(order.orderId)
	} else {
		orderBook.SellList.DeleteOrder(order.orderId)
	}
}

// ExecuteOrder executes an order on the order book
// It tries to match a new order with the order book and executes if there is a match.
// If no match, the new order is added to the order book.
func (ob *OrderBooks) ExecuteOrder(newOrder Order, store *Store) {
	orderBook := ob.getOrderBook(newOrder.assetId)
	orderBook.Lock()
	defer orderBook.Unlock()
	if newOrder.buyOrSell == BUY {
		sellList := orderBook.SellList
		buyOrder := executeOrder(sellList, newOrder, store, BUY)

		// add unfilled buy orders to the order book
		if buyOrder.size > 0 {
			orderBook.BuyList.AddOrder(buyOrder)
		}
	} else {
		buyList := orderBook.BuyList
		sellOrder := executeOrder(buyList, newOrder, store, SELL)

		// add unfilled sell orders to the order book
		if sellOrder.size > 0 {
			orderBook.SellList.AddOrder(sellOrder)
		}
	}
}

// executeOrder tries to execute an order if a match order is found
// else adds the order to the order book
func executeOrder(orderList *OrdersList, newOrder Order, store *Store, buyOrSell BuyOrSell) Order {
	for orderMatchAvailable(orderList, newOrder, buyOrSell) { // match incoming order with orders in the order book
		matchedOrder := orderList.GetTopOrder()
		matchedPrice := getMatchedPrice(buyOrSell, matchedOrder, newOrder)

		tradeAssetsSize := min(matchedOrder.size, newOrder.size)
		newOrder.size -= tradeAssetsSize     // update new order's asset size
		matchedOrder.size -= tradeAssetsSize // update matched order in order book

		// new order completely filled
		if newOrder.size == 0 {
			// matchedOrder and newOrder both completely filled
			if matchedOrder.size == 0 {
				orderList.DeleteOrder(matchedOrder.orderId)
				// update matched order user's asset info in store
				store.UpdateUserAsset(matchedOrder, matchedPrice, tradeAssetsSize, buyOrSell, Complete)

			} else {
				orderList.UpdateOrder(matchedOrder)

				// update matched order user's asset info in store
				store.UpdateUserAsset(matchedOrder, matchedPrice, tradeAssetsSize, buyOrSell, Working)
			}

			// update new order user's assets info in store
			store.UpdateUserAsset(newOrder, matchedPrice, tradeAssetsSize, buyOrSell, Complete)

			break // exit loop since new order was fulfilled
		}

		// matched order completely executed
		// remove matched order from order book
		orderList.DeleteOrder(matchedOrder.orderId)

		// update new order user's assets info in store
		store.UpdateUserAsset(newOrder, matchedPrice, tradeAssetsSize, buyOrSell, Working)

		// update matched order user's asset info in store
		store.UpdateUserAsset(matchedOrder, matchedPrice, tradeAssetsSize, buyOrSell, Complete)
	}
	return newOrder
}

// getMatchedPrice returns matched price, which is the price of the sell order.
func getMatchedPrice(buyOrSell BuyOrSell, matchedOrder Order, newOrder Order) Usd {
	var matchedPrice Usd
	if buyOrSell == BUY {
		matchedPrice = matchedOrder.limit
	} else {
		matchedPrice = newOrder.limit
	}
	return matchedPrice
}

// orderMatchAvailable returns if there is an order in the order book that matches the incoming new order
// For a BUY order, it returns true if there is a sell order in the order book that is <= the buy order's limit price
// For a sell order, it returns true if there is a buy order in the order book that is >= the sell order's limit price
// Else returns false.
func orderMatchAvailable(orderList *OrdersList, newOrder Order, orderType BuyOrSell) bool {
	return (!orderList.isEmpty() && orderType == BUY && orderList.GetTopOrder().limit <= newOrder.limit) ||
		(!orderList.isEmpty() && orderType == SELL && orderList.GetTopOrder().limit >= newOrder.limit)
}
