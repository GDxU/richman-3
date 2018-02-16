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

	// get my Balance ? why?
	go func() {
		for {
			myAccounts = account.GetBalance()
			time.Sleep(time.Duration(2) * time.Second)
		}
	}()

	time.Sleep(time.Duration(1) * time.Second)

	mlo := myAccounts.GetLimitOrders("BTC")
	fmt.Println(mlo)

	fmt.Println(myAccounts.BuyCoin("BTC", 100000, 1.0))

	mlo = myAccounts.GetLimitOrders("BTC")
	fmt.Println(mlo)

	limitOrder := mlo.LimitOrders[0]
	p, _ := strconv.Atoi(limitOrder.Price)
	q, _ := strconv.ParseFloat(limitOrder.Qty, 64)

	s := myAccounts.CancelOrder(limitOrder.OrderId, uint64(p), q, limitOrder.Type)
	fmt.Println(s)

	mlo = myAccounts.GetLimitOrders("BTC")
	fmt.Println(mlo)

	fmt.Scanln()
}
