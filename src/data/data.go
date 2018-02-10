package data

import (
	"database/sql"
	"dbfunc"
	"encoding/json"
	"fmt"
	"logger"
	"net/http"
	"strconv"
	"time"
)

const BaseURL string = "https://api.coinone.co.kr"

type ResBody struct {
	ErrorCode      string
	Timestamp      string
	CompleteOrders []struct {
		Timestamp string
		Price     string
		Qty       string
	}
}

func GetCoinData(s string, duration int, db *sql.DB) {
	logger := logger.GetLogger("[Get " + s + " Data]")
	url := BaseURL + "/trades?currency=" + s
	for {
		go func() {
			res, err := http.Get(url)
			if err != nil {
				fmt.Print(err)
			} else {
				resbody := ResBody{}
				err2 := json.NewDecoder(res.Body).Decode(&resbody)
				if err2 == nil {
					price := resbody.refine()
					price.Insert(db)
				} else {
					fmt.Print(err2)
				}
			}
		}()
		logger.Println("Get Data Succeeded")
		time.Sleep(time.Duration(duration*60) * time.Second)
	}
}

func (r *ResBody) refine() *dbfunc.Price {
	total := 0.0
	price := new(dbfunc.Price)
	lastOrder := len(r.CompleteOrders)
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
	return price
}
