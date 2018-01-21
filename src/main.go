package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"
	"dbfunc"
	"database/sql"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
)

type CONSTANTS struct {
	ACCESS_TOKEN string
	SECRET_KEY string
	BASE_URL string
}

type PAYLOAD struct {
	Currency string
	Period string
}

type RESBODY struct {
	ErrorCode string
	Timestamp string
	CompleteOrders[] struct {
		Timestamp string
		Price string
		Qty string
	}
}

func main() {
	constants := CONSTANTS{
		ACCESS_TOKEN : "6ef090a1-ad27-4123-afe7-7743e84c2231",
		SECRET_KEY : "a21cf668-cd9f-4be3-abc8-c66580ceb813",
		BASE_URL : "https://api.coinone.co.kr",
	}
	db, err := sql.Open("mysql", dbfunc.DBUSER+":"+dbfunc.DBAUTH+"@tcp("+dbfunc.DBIPADDR+":"+dbfunc.DBPORT+")/btc")
	err2 := db.Ping()	
	if err != nil {
		panic(err.Error)
	} else if err2 != nil {
		fmt.Println(err2.Error())
		panic(err2.Error)
	} else {
		fmt.Print("DB Connected.")
	}

	go constants.getCoinData("BTC", 10, db)
	fmt.Scanln()
}

func (c *CONSTANTS) getCoinData(s string, duration int, db *sql.DB){
	url := c.BASE_URL + "/trades"
	for{
		res, err := http.Get(url)
		if err != nil {
			fmt.Print(err)
		} else {
			resbody := RESBODY{}
			err2 := json.NewDecoder(res.Body).Decode(&resbody)
			if err2 == nil {
				price := resbody.refine()
				price.Insert(db);
			} else {
				fmt.Print(err2)
			}
		}
		time.Sleep(time.Duration(duration * 60) * time.Second)
	}
}

func (r *RESBODY) refine() *dbfunc.PRICE {
	total := 0.0
	price := new(dbfunc.PRICE)
	lastOrder := len(r.CompleteOrders)
	for i := lastOrder-1 ; ; i-- {
		co := r.CompleteOrders[i]
		qty, _ := strconv.ParseFloat(co.Qty, 64)
		ts, _ := strconv.ParseUint(co.Timestamp, 10, 64)
		total = total + qty
		if i == lastOrder -1  {
			price.Timestamp, _= strconv.ParseUint(co.Timestamp, 10, 64)
			price.LastPrice, _ = strconv.ParseUint(co.Price, 10, 64)
			price.MaxPrice = price.LastPrice
			price.MinPrice = price.LastPrice
			fmt.Println(price.Timestamp)
		} else if ts < (price.Timestamp - 600) {
			price.FirstPrice, _ = strconv.ParseUint(co.Price, 10, 64)
			fmt.Println(ts)
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
