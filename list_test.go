package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrdersList_AddOrder_BUY(t *testing.T) {
	o1 := Order{orderId: "1", limit: 100, buyOrSell: BUY}
	o2 := Order{orderId: "2", limit: 200, buyOrSell: BUY}
	o3 := Order{orderId: "3", limit: 300, buyOrSell: BUY}
	o4 := Order{orderId: "4", limit: 400, buyOrSell: BUY}

	list := newOrdersList()
	list.AddOrder(o2) // add to front
	list.AddOrder(o4) // add to front
	list.AddOrder(o3) // add to mid
	list.AddOrder(o1) // add to end

	// assert correct order of orders in list
	tracker := list.front
	assert.Equal(t, o4, tracker.val)
	tracker = tracker.next
	assert.Equal(t, o3, tracker.val)
	tracker = tracker.next
	assert.Equal(t, o2, tracker.val)
	tracker = tracker.next
	assert.Equal(t, o1, tracker.val)
}

func TestOrdersList_AddOrder_SELL(t *testing.T) {
	o1 := Order{orderId: "1", limit: 100, buyOrSell: SELL}
	o2 := Order{orderId: "2", limit: 200, buyOrSell: SELL}
	o3 := Order{orderId: "3", limit: 300, buyOrSell: SELL}
	o4 := Order{orderId: "4", limit: 400, buyOrSell: SELL}

	list := newOrdersList()
	list.AddOrder(o2) // add to front
	list.AddOrder(o4) // add to end
	list.AddOrder(o3) // add to mid
	list.AddOrder(o1) // add to front

	// assert correct order of sell orders in list
	tracker := list.front
	assert.Equal(t, o1, tracker.val)
	tracker = tracker.next
	assert.Equal(t, o2, tracker.val)
	tracker = tracker.next
	assert.Equal(t, o3, tracker.val)
	tracker = tracker.next
	assert.Equal(t, o4, tracker.val)
}

func TestOrdersList_GetTopOrder(t *testing.T) {
	o1 := Order{orderId: "1", limit: 100}
	o2 := Order{orderId: "2", limit: 200}

	list := newOrdersList()
	list.AddOrder(o1)
	list.AddOrder(o2)

	assert.Equal(t, o2, list.front.val)
	assert.Equal(t, o2, list.GetTopOrder())
}

func TestOrdersList_UpdateOrder(t *testing.T) {
	o1 := Order{orderId: "1", limit: 100, size: 10}
	o2 := Order{orderId: "2", limit: 200, size: 20}
	o3 := Order{orderId: "3", limit: 300, size: 30}


	list := newOrdersList()
	list.AddOrder(o1)
	list.AddOrder(o2)
	list.AddOrder(o3)

	// update order
	o2.size= 50
	list.UpdateOrder(o2)

	actual := list.getOrder(o2.orderId)
	assert.Equal(t, 50, actual.size)
}

func TestOrdersList_DeleteOrder(t *testing.T) {
	o1 := Order{orderId: "1", limit: 100, size: 10}
	o2 := Order{orderId: "2", limit: 200, size: 20}
	o3 := Order{orderId: "3", limit: 300, size: 30}


	list := newOrdersList()
	list.AddOrder(o1)
	list.AddOrder(o2)
	list.AddOrder(o3)

	// delete order -> mid of list
	list.DeleteOrder(o2.orderId)

	actual := list.getOrder(o2.orderId)
	assert.Empty(t,  actual)
	assert.Equal(t, 2, list.getSize())

	// delete order -> end of list
	list.DeleteOrder(o3.orderId)
	actual = list.getOrder(o3.orderId)

	assert.Empty(t,  actual)
	assert.Equal(t, 1, list.getSize())

	// delete order -> front of list
	list.DeleteOrder(o1.orderId)
	actual = list.getOrder(o1.orderId)

	assert.Empty(t,  actual)
	assert.Equal(t, 0, list.getSize())
	assert.Nil(t, list.front)
}