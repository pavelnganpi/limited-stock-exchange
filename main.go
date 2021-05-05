package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main()  {

	s := newOrderMatchingService()
	defer s.Close()
	r := mux.NewRouter()

	r.HandleFunc("/users", s.InitExchangeHandler).Methods("POST")
	r.HandleFunc("/users/{userId}/orders", s.CreateOrderHandler).Methods("POST")
	r.HandleFunc("/users/{userId}/orders/{orderId}", s.CancelOrderHandler).Methods("DELETE")
	r.HandleFunc("/users/{userId}/orders", s.GetOrdersHandler).Methods("GET")


	log.Fatal(http.ListenAndServe("0.0.0.0:9093", r))
}
