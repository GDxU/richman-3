package main

import (
	"account"
	"database/sql"
	"dbfunc"
	"encoding/json"
	"fmt"
	"log"
	"logger"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const BaseURL string = "https://api.coinone.co.kr"

type PayLoad struct {
	Currency string
	Period   string
}

type ResBody struct {
	ErrorCode      string
	Timestamp      string
	CompleteOrders []struct {
		Timestamp string
		Price     string
		Qty       string
	}
}

func main() {
	logger.Now = time.Now().Format(time.RFC822)
	logger := logger.GetLogger("[Let's get Rich]")

	db, err := sql.Open("mysql", dbfunc.DBUSER+":"+dbfunc.DBAUTH+"@tcp("+dbfunc.DBIPADDR+":"+dbfunc.DBPORT+")/btc")
	err2 := db.Ping()
	if err != nil {
		panic(err.Error)
	} else if err2 != nil {
		fmt.Println(err2.Error())
		panic(err2.Error)
	} else {
		logger.Print("DB Connected.")
	}
	go func() {
		for {
			time.Sleep(time.Duration(2) * time.Second)
			account.GetBalance()
		}
	}()
	fmt.Scanln()
	//getCoinData("BTC", 10, db, logger)
}

func getCoinData(s string, duration int, db *sql.DB, logger *log.Logger) {
	url := BaseURL + "/trades"
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
