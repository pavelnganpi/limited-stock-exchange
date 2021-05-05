package main

// OrderMatchingService manages order matching executes trades for buy and sell limit orders
type OrderMatchingService struct {
	Store      *Store        // in memory data db
	OrderBooks *OrderBooks   // order book for each asset
	OCh        chan OrderReq // channel to process incoming orders synchronously
}

func newOrderMatchingService() *OrderMatchingService {
	s := &OrderMatchingService{
		Store:      newStore(),
		OrderBooks: newOrderBooks(),
		OCh:        make(chan OrderReq, 100),
	}

	go s.ProcessOrderReqs() // process orders in a goroutine(process) independently

	return s
}

// InitExchange initializes the exchange with users and their assets
func (s *OrderMatchingService) InitExchange(reqs []InitExchangeReq) {
	for _, r := range reqs {
		s.Store.CreateUser(r)
	}
}

// ProcessOrderReqs processes new orders, attempts to execute an order if is there is a match
// if not adds the order to the order book
func (s *OrderMatchingService) ProcessOrderReqs() {
	for or := range s.OCh {
		order := createOrderFromOrderReq(or)
		s.SaveOrderToStore(order) // save new order to db
		s.ExecuteOrder(order)
	}
}

// GetUserActiveOrders returns user's active orders
func (s *OrderMatchingService) GetUserActiveOrders(userId UserId) []OrderResp {
	var activeOrders []OrderResp
	for _, order := range s.Store.GetUserData(userId).orders {
		if order.status == Working {
			activeOrders = append(activeOrders, orderToOrderResp(order))
		}
	}

	return activeOrders
}

// GetUserCompleteOrders returns a user's complete orders
func (s *OrderMatchingService) GetUserCompleteOrders(userId UserId) []OrderResp {
	var completeOrders []OrderResp
	for _, order := range s.Store.GetUserData(userId).orders {
		if order.status == Complete {
			completeOrders = append(completeOrders, orderToOrderResp(order))
		}
	}

	return completeOrders
}

// CancelUserOrder cancels a user's order
func (s *OrderMatchingService) CancelUserOrder(userId UserId, orderId OrderId) Order {
	if s.Store.GetUserData(userId).orders != nil {
		order := s.Store.GetUserData(userId).orders[orderId]
		s.OrderBooks.DeleteOrder(order)                                            // remove order from order book
		s.Store.UpdateUserAssetOnOrderCancel(userId, order.assetId, order.orderId) // update order status to cancel
	}
	return Order{}
}

// SaveOrderToStore stores an order in the store(db)
func (s *OrderMatchingService) SaveOrderToStore(order Order) { // TODO add delete order from db
	s.Store.AddUserOrder(order)
}

// ExecuteOrder tries to execute an order if a match order is found
// else adds the order to the order book
func (s *OrderMatchingService) ExecuteOrder(order Order) {
	//fmt.Println("EXEC", order)
	s.OrderBooks.ExecuteOrder(order, s.Store)
}

// Close closes the channel
func (s *OrderMatchingService) Close() {
	close(s.OCh)
}
