# limited-stock-exchange

Simple representation of how a stock exchange works with limit orders for buying and selling assets.

To run the app locally, make sure you have docker installed and running, Follow below steps to run.

1. `git clone git@github.com:paveyn/limited-stock-exchange.git`
2. `cd limited-stock-exchange`
3. `go test` to run the unit tests
4. `docker build -t limited-stockexchange-app .`
5. `docker run -p 9093:9093 -it limited-stockexchange-app`

App should be running locally on `locahost:9093`

Endpoints

1. `Post /users` to initialise the stock exchange with some users and assets. E.g
```
curl -X "POST" "http://localhost:9093/users" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'[
  {
    "user_id": "user1",
    "cash": 100000,
    "assets": [
      {
        "asset_id": "COIN",
        "size": 100
      },
      {
        "asset_id": "GAME",
        "size": 200
      },
      {
        "asset_id": "AAPL",
        "size": 500
      }
    ]
  },
  {
    "user_id": "user2",
    "cash": 200000,
    "assets": [
      {
        "asset_id": "COIN",
        "size": 100
      },
      {
        "asset_id": "GAME",
        "size": 700
      },
      {
        "asset_id": "AAPL",
        "size": 1500
      }
    ]
  }
]'
```
2. `Post /users/{:userId}/orders` to create an order for a user. E.g
```
curl -X "POST" "http://localhost:9093/users/user1/orders" \
     -H 'Content-Type: application/json' \
     -d $'{
  "asset_id": "COIN",
  "buy_or_sell": 0,
  "size": 10,
  "limit": 100
}'
```
3. `Delete /users/{:userId}/orders/{:orderId}` to cancel user's order. E.g
```
curl -X "DELETE" "http://localhost:9093/users/user1/orders/aEWEjxa3sCshvacGNChtcn" \
     -H 'Content-Type: application/json' \
     -d $'{
  "value": 200
}'
```
4. `Get /users/{:userId}/orders?status={order status}` to get a user order status. `status=active` for active orders, `status=complete` for completed orders. E.g
```
curl "http://localhost:9093/users/user1/orders?status=active" \
     -H 'Content-Type: application/json' \
     -d $'{
  "value": 200
}'
curl "http://localhost:9093/users/user1/orders?status=complete" \
     -H 'Content-Type: application/json' \
     -d $'{
  "value": 200
}'
```
