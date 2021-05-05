package main

import (
	"time"
)

// type definitions
type UserId string
type AssetId string
type OrderId string
type BuyOrSell int
type Usd int // in cents
type OrderStatus string

// enums
const (
	BUY  BuyOrSell = iota // 0
	SELL                  // 1
)

const (
	Working  OrderStatus = "WORKING"
	Complete OrderStatus = "COMPLETE"
	Canceled OrderStatus = "CANCELED"
)

// Order struct represents an order
type Order struct {
	orderId   OrderId     // id of val
	userId    UserId      // id of user who owns the val
	limit     Usd         // limit price, in Usd cents
	assetId   AssetId     // asset to trade
	size      int         // number of assets
	buyOrSell BuyOrSell   // buy or sell val
	eventAt   time.Time   // time when val was created
	status    OrderStatus // status of the val
	filled    int         // total number of assets filled during a trade
}

// UserData struct represents a struct for storing user assets and orders
type UserData struct {
	userId UserId
	cash   Usd               // cash amount in Usd e.g $100 -> 10000 Usd
	assets map[AssetId]int   // map of AssetId -> size of asset
	orders map[OrderId]Order // map of OrderId -> val metadata
}

// Store acts the database. An in memory db
type Store struct {
	db map[UserId]UserData
}

func newStore() *Store {
	return &Store{
		db: make(map[UserId]UserData),
	}
}

// CreateUser crates a new user for the exchange
func (s *Store) CreateUser(req InitExchangeReq) {
	userData := UserData{
		userId: req.UserId,
		cash:   req.Cash,
		assets: make(map[AssetId]int),   // init assets map for every user
		orders: make(map[OrderId]Order), // init orders map for every user
	}
	for _, asset := range req.Assets {
		userData.assets[asset.AssetId] = asset.Size
	}
	s.db[req.UserId] = userData
}

// GetUserData gets a user's data
func (s *Store) GetUserData(userId UserId) UserData {
	return s.db[userId]
}

// AddUserOrder adds an order to a user's data.
func (s *Store) AddUserOrder(order Order) {
	userData := s.GetUserData(order.userId)
	userData.orders[order.orderId] = order
	if order.buyOrSell == BUY { // decrease user's available cash on every new buy order created
		newCash := userData.cash - getTotalAssetCost(order.limit, order.size)
		userData.cash = newCash
	} else { // decrease user's asset size on every new sell order
		userData.assets[order.assetId] -= order.size
	}

	s.db[order.userId] = userData
}

// UpdateUserAssetOnSuccessBuy updates a user's assets size and order status upon a success buy event
func (s *Store) UpdateUserAssetOnSuccessBuy(userId UserId, assetId AssetId, orderId OrderId, tradeAssetSize int, status OrderStatus) {
	userData := s.GetUserData(userId)
	userData.assets[assetId] += tradeAssetSize // increase asset size for newly bought asset

	order := userData.orders[orderId]
	order.status = status
	order.filled += tradeAssetSize

	userData.orders[orderId] = order

	s.db[userId] = userData
}

// UpdateUserAssetOnSuccessSell updates a user's assets available cash and order status upon a success sale event
func (s *Store) UpdateUserAssetOnSuccessSell(userId UserId, orderId OrderId, cashGain Usd, status OrderStatus, tradeAssetSize int) {
	userData := s.GetUserData(userId)
	userData.cash += cashGain // increase cash available after newly sold asset

	order := userData.orders[orderId]
	order.status = status
	order.filled += tradeAssetSize

	userData.orders[orderId] = order

	s.db[userId] = userData
}

// UpdateUserAsset updates a user's assets upon a buy or sell event
func (s *Store) UpdateUserAsset(order Order, matchedPrice Usd, tradeAssetsSize int, orderType BuyOrSell, status OrderStatus) {
	if (orderType == BUY && order.buyOrSell == SELL) || (orderType == SELL && order.buyOrSell == SELL) {
		s.UpdateUserAssetOnSuccessSell(order.userId, order.orderId, getTotalAssetCost(matchedPrice, tradeAssetsSize), status, tradeAssetsSize)
	} else {
		s.UpdateUserAssetOnSuccessBuy(order.userId, order.assetId, order.orderId, tradeAssetsSize, status)
	}
}

// UpdateUserAssetOnOrderCancel updates a user's order status open a cancel order event
func (s *Store) UpdateUserAssetOnOrderCancel(userId UserId, assetId AssetId, orderId OrderId) {
	userData := s.GetUserData(userId)
	order := userData.orders[orderId]

	// if order is a buy order, reallocate back cash deducted from buy order
	if order.buyOrSell == BUY {
		userData.cash += getTotalAssetCost(order.limit, order.size)
		// else sell order, reallocate back asset size deducted from sell order
	} else {
		userData.assets[assetId] += order.size
	}

	order.status = Canceled // mark order as canceled
	userData.orders[orderId] = order

	s.db[userId] = userData
}
