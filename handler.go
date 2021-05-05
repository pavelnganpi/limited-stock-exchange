package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type InitExchangeReq struct {
	UserId UserId  `json:"user_id"`
	Assets []Asset `json:"assets"`
	Cash   Usd     `json:"cash"`
}

type Asset struct {
	AssetId AssetId `json:"asset_id"`
	Size    int     `json:"size"`
}

type OrderReq struct {
	UserId    UserId    `json:"user_id"`     // id of user making the val
	Limit     Usd       `json:"limit"`       // Limit price, in usd cents
	AssetId   AssetId   `json:"asset_id"`    // asset to trade
	Size      int       `json:"size"`        // number of assets
	BuyOrSell BuyOrSell `json:"buy_or_sell"` // buy or sell val
}

type OrderResp struct {
	OrderId   OrderId     `json:"order_id"`    // id of val
	UserId    UserId      `json:"user_id"`     // id of user who owns the val
	Limit     Usd         `json:"limit"`       // Limit price, in Usd cents
	AssetId   AssetId     `json:"asset_id"`    // asset to trade
	Size      int         `json:"size"`        // number of assets
	BuyOrSell BuyOrSell   `json:"buy_or_sell"` // buy or sell val
	EventAt   time.Time   `json:"event_at"`    // time when val was created
	Status    OrderStatus `json:"status"`      // Status of the val
	Filled    int         `json:"filled"`      // total number of assets filled during a trade
}

// InitExchangeHandler handles requests to initialize the stock exchange with users and their assets
func (s *OrderMatchingService) InitExchangeHandler(w http.ResponseWriter, r *http.Request) {
	var req []InitExchangeReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		println(err.Error())
		return
	}

	s.InitExchange(req)
	JSONResponse(w, http.StatusOK, struct{}{})
}

// CreateOrderHandler handles request to process buy and sell orders
func (s *OrderMatchingService) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var or OrderReq
	err := json.NewDecoder(r.Body).Decode(&or)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userId := mux.Vars(r)["userId"]
	or.UserId = UserId(userId)

	err = validateOrderReq(s.Store.GetUserData(UserId(userId)), or)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.OCh <- or

	JSONResponse(w, http.StatusOK, struct{}{})
}

// CancelOrderHandler handles request to cancel order
func (s *OrderMatchingService) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["userId"]
	orderId := mux.Vars(r)["orderId"]

	JSONResponse(w, http.StatusNoContent, s.CancelUserOrder(UserId(userId), OrderId(orderId)))
}

// GetOrdersHandler handles request to get all user orders by status
func (s *OrderMatchingService) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["userId"]
	status := r.URL.Query().Get("status")
	var resp []OrderResp
	if status == "complete" {
		resp = s.GetUserCompleteOrders(UserId(userId))
		// return active orders by default
	} else {
		resp = s.GetUserActiveOrders(UserId(userId))
	}

	JSONResponse(w, http.StatusOK, resp)
}

func JSONResponse(w http.ResponseWriter, code int, output interface{}) {
	response, _ := json.Marshal(output)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
