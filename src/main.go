package main

import (
	"account"
	"data"
	"dbfunc"
	"fmt"
	"logger"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	logger.Now = time.Now().Format(time.RFC822)
	logger := logger.GetLogger("[Let's get Rich]")
	logger.Println("Let's Get Start!")

	myAccounts := new(account.MyBalance)
	db := dbfunc.GetDbConn("BTC")

	// get BTC Trade data every 10 minutes.
	go func() {
		for {
			data.GetCoinTradeData("BTC", db)
			time.Sleep(time.Duration(10) * time.Minute)
		}
	}()

	go func() {
		for {
			ctp := dbfunc.Select(db, "BTC", 5)
			var tangent float64 = (ctp[0].Bolband - ctp[1].Bolband) / ctp[0].Bolband
			ro := data.GetRecentOrder("BTC")
			currentValue, _ := strconv.ParseUint(ro.Ask.Price, 10, 64)
			if currentValue < (ctp[0].Bolband - 5*uint64(ctp[0].Bolbandsd)/2) {

			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()

	fmt.Scanln()
}
