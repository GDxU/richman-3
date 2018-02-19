package data

import (
	"account"
	"database/sql"
	"dbfunc"
	"encoding/json"
	"logger"
	"net/http"
	"strconv"
	"time"
)

// ResBody is Response Body of GetTradeData
type ResBody struct {
	ErrorCode      string
	Timestamp      string
	CompleteOrders []struct {
		Timestamp string
		Price     string
		Qty       string
	}
}

// OrderBook is current OrderBook
type OrderBook struct {
	ErrorCode string
	Timestamp string
	Currency  string
	Ask       []struct {
		Price string
		Qty   string
	}
	Bid []struct {
		Price string
		Qty   string
	}
}

// RecentOrder is Recent Order Book
type RecentOrderBook struct {
	ErrorCode string
	Timestamp string
	Currency  string
	Ask       struct {
		Price string
		Qty   string
	}
	Bid struct {
		Price string
		Qty   string
	}
}

// GetOneDayTradeData gets last one day's trade data.
// @param: coin name like "BTC"
// @param: *sql.DB
func GetOneDayTradeData(coin string, db *sql.DB) {
	logger := logger.GetLogger("[Get One Day " + coin + " Trade Data]")
	url := account.BaseURL + "/trades?currency=" + coin + "&period=" + "day"

	res, err := http.Get(url)
	if err != nil {
		logger.Println(err)
	} else {
		resbody := ResBody{}
		err2 := json.NewDecoder(res.Body).Decode(&resbody)
		if err2 == nil {
			prices := resbody.refine(true)
			if prices != nil {
				for i := len(prices) - 1; i >= 0; i-- {
					prices[i].Insert(db, coin)
				}
				logger.Println("Get One Day Coin Data")
			}
		}
	}
}

// GetCoinTradeData gets Trade Data of a coin from CoinOne
// @param: coin name like "BTC"
// @param: *sql.DB
func GetCoinTradeData(coin string, db *sql.DB) {
	logger := logger.GetLogger("[Get " + coin + " Trade Data]" + time.Now().Format(time.RFC3339))
	url := account.BaseURL + "/trades?currency=" + coin

	res, err := http.Get(url)
	if err != nil {
		logger.Println(err)
	} else {
		resbody := ResBody{}
		err2 := json.NewDecoder(res.Body).Decode(&resbody)
		if err2 == nil {
			prices := resbody.refine(false)
			if prices != nil {
				prices[0].Insert(db, coin)
			}
		} else {
			logger.Println(err2)
		}
	}

	logger.Println("Get Data Succeeded")
}

// GetRecentOrder returns current OrderBook
// @return: an OrderBook, or nil if err
func GetRecentOrder(coin string) *RecentOrderBook {
	logger := logger.GetLogger("[Get " + coin + " OrderBook]")
	url := account.BaseURL + "/orderbook/?currency=" + coin

	res, err := http.Get(url)
	orderBook := new(OrderBook)
	if err != nil {
		logger.Println(err)
		return nil
	}
	err2 := json.NewDecoder(res.Body).Decode(orderBook)
	if err2 != nil {
		logger.Println(err2)
		return nil
	}
	recentOrder := new(RecentOrderBook)
	recentOrder.Ask = orderBook.Ask[0]
	recentOrder.Bid = orderBook.Bid[0]
	recentOrder.Currency = orderBook.Currency
	recentOrder.Timestamp = orderBook.Timestamp
	recentOrder.ErrorCode = orderBook.ErrorCode

	return recentOrder
}

func (r *ResBody) refine(isAll bool) []dbfunc.CoinTradePrice {
	prices := []dbfunc.CoinTradePrice{}
	lastOrder := len(r.CompleteOrders)
	if lastOrder <= 0 {
		return nil
	}
	if isAll {
		total := 0.0
		price := new(dbfunc.CoinTradePrice)
		tempTimestamp, _ := strconv.ParseUint(r.Timestamp, 10, 64)
		for i := lastOrder - 1; i >= 0; i-- {
			co := r.CompleteOrders[i]
			qty, _ := strconv.ParseFloat(co.Qty, 64)
			ts, _ := strconv.ParseUint(co.Timestamp, 10, 64)
			total = total + qty
			if price.Timestamp2 == 0 {
				price.Timestamp2, _ = strconv.ParseUint(co.Timestamp, 10, 64)
				price.LastPrice, _ = strconv.ParseUint(co.Price, 10, 64)
				price.MaxPrice = price.LastPrice
				price.MinPrice = price.LastPrice
			} else if ts < tempTimestamp-600 {
				price.FirstPrice, _ = strconv.ParseUint(co.Price, 10, 64)
				price.Timestamp1, _ = strconv.ParseUint(co.Timestamp, 10, 64)
				price.AvgPrice = (price.MaxPrice + price.MinPrice) / 2
				price.Qty = total
				prices = append(prices, *price)
				tempTimestamp = tempTimestamp - 600
				price = new(dbfunc.CoinTradePrice)
				total = 0.0
			} else if i == 0 {
				price.FirstPrice, _ = strconv.ParseUint(co.Price, 10, 64)
				price.Timestamp1, _ = strconv.ParseUint(co.Timestamp, 10, 64)
				price.AvgPrice = (price.MaxPrice + price.MinPrice) / 2
				price.Qty = total
				prices = append(prices, *price)
				break
			} else {
				curPrice, _ := strconv.ParseUint(co.Price, 10, 64)
				if curPrice > price.MaxPrice {
					price.MaxPrice = curPrice
				}
				if curPrice < price.MinPrice {
					price.MinPrice = curPrice
				}
			}
		}
	} else {
		total := 0.0
		price := new(dbfunc.CoinTradePrice)
		for i := lastOrder - 1; ; i-- {
			co := r.CompleteOrders[i]
			qty, _ := strconv.ParseFloat(co.Qty, 64)
			ts, _ := strconv.ParseUint(co.Timestamp, 10, 64)
			total = total + qty
			if i == lastOrder-1 {
				price.Timestamp2, _ = strconv.ParseUint(co.Timestamp, 10, 64)
				price.LastPrice, _ = strconv.ParseUint(co.Price, 10, 64)
				price.MaxPrice = price.LastPrice
				price.MinPrice = price.LastPrice
			} else if ts < (price.Timestamp2 - 600) {
				price.FirstPrice, _ = strconv.ParseUint(co.Price, 10, 64)
				price.Timestamp1, _ = strconv.ParseUint(co.Timestamp, 10, 64)
				break
			} else {
				curPrice, _ := strconv.ParseUint(co.Price, 10, 64)
				if curPrice > price.MaxPrice {
					price.MaxPrice = curPrice
				}
				if curPrice < price.MinPrice {
					price.MinPrice = curPrice
				}
			}
		}
		price.AvgPrice = (price.MaxPrice + price.MinPrice) / 2
		price.Qty = total
		prices = append(prices, *price)
	}
	return prices
}
