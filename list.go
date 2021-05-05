package main

import (
	"fmt"
)

// Chose to use a linked list data structure to order the order book over an array based list due to following reasons
//
// 1. I am assuming the best case in an ideal world, users will make limit orders close to the price at the top of the order book
// 	hence using a list based data structure works best over a tree base data structure(heaps, balanced search tree) due to O(1) time complexity over Log(N) since
//	one has to just scan from start of the list, looking for an appropriate position to insert an order in the order book.
//
// 2. Due to how array copy works in golang, inserting a new order will be O(N), since golang will have to shift all elements after i+1 to insert the new order
// 3. Chose a linked list approach since insert work in place and no need to copy over data
// 4. In best case(order close to the start of the list), deletes, insert, updates and get best order work at ~O(1) since the best order will always be at the start of the list

type OrderNode struct {
	val  Order
	next *OrderNode
}

// OrdersList keeps a linked list of OrderNode{} in price-time priority
type OrdersList struct {
	front *OrderNode
}

func newOrdersList() *OrdersList {
	return &OrdersList{}
}

// AddOrder adds an order to the list, maintains order of price-time priority
// Since this works in first come first serve, time priority is automatically maintained.
// e.g [4,2,1], if another 2 comes in at a later time, it will be inserted before 1, -> [4,2,2,1].
func (l *OrdersList) AddOrder(newOrder Order) {
	newOrderNode := &OrderNode{
		val: newOrder,
	}
	if l.front == nil {
		l.front = newOrderNode
	} else {

		// if buy order, order order book by highest buy price(limit)-time priority
		if newOrder.buyOrSell == BUY {
			// check if new order is better than that at the front of the list
			// if so set new front of list to new order
			if newOrder.limit > l.front.val.limit {
				temp := l.front
				l.front = newOrderNode
				newOrderNode.next = temp
				return
			}

			// find best position to insert order by price-time priority
			// insert newOrder
			for t := l.front; t != nil; t = t.next {
				if (t.next != nil && newOrder.limit > t.next.val.limit) || t.next == nil {
					temp := t.next
					t.next = newOrderNode
					newOrderNode.next = temp
					break // exit once new order has been inserted
				}
			}

			// do reverse for sell orders. sell order with lowest price should be at the head
		} else {
			if newOrder.limit < l.front.val.limit {
				temp := l.front
				l.front = newOrderNode
				newOrderNode.next = temp
				return
			}

			// find best position to insert order by price-time priority
			// insert newOrder
			for t := l.front; t != nil; t = t.next {
				if (t.next != nil && newOrder.limit < t.next.val.limit) || t.next == nil {
					temp := t.next
					t.next = newOrderNode
					newOrderNode.next = temp
					break // exit once new order has been inserted
				}
			}
		}
	}
}

// GetTopOrder returns the top order in the list, which is the front of the list
func (l *OrdersList) GetTopOrder() Order {
	if l.front != nil {
		return l.front.val
	}
	return Order{}
}

// DeleteOrder deletes an order from the list given an order id
func (l *OrdersList) DeleteOrder(oid OrderId) {
	// handle case order to delete at at the front of the list
	if l.front != nil {
		if l.front.val.orderId == oid {
			l.front = l.front.next
			return
		}

		for t := l.front; t != nil; t = t.next {
			if t.next != nil && t.next.val.orderId == oid {
				temp := t.next.next
				t.next = temp
			}
		}
	}
}

// UpdateOrder updates an order in the list. Only the size of the order can be updated.
func (l *OrdersList) UpdateOrder(order Order) {
	for t := l.front; t != nil; t = t.next {
		if t.val.orderId == order.orderId {
			t.val.size = order.size
		}
	}
}

// getOrder finds an Order{} given an order id and returns the Order{} if found.
func (l *OrdersList) getOrder(oid OrderId) Order {
	for t := l.front; t != nil; t = t.next {
		if t.val.orderId == oid {
			return t.val
		}
	}
	return Order{}
}

// getSize returns the size of the list
func (l *OrdersList) getSize() int {
	count := 0
	for t := l.front; t != nil; t = t.next {
		count++
	}
	return count
}

func (l *OrdersList) isEmpty() bool {
	return l.front == nil
}

func (l *OrdersList) print() {
	for t := l.front; t != nil; t = t.next {
		fmt.Println(t)
	}
}
