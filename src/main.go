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

	myAccounts := account.GetBalance()
	myLimitOrders := myAccounts.GetLimitOrders("BTC")
	db := dbfunc.GetDbConn("BTC")

	// get Account Info every 10 seconds
	go func() {
		myAccounts = account.GetBalance()
		myLimitOrders = myAccounts.GetLimitOrders("BTC")
		time.Sleep(time.Duration(10) * time.Second)
	}()

	data.GetOneDayTradeData("BTC", db)
	// get BTC Trade data every 10 minutes.
	go func() {
		for {
			time.Sleep(time.Duration(10) * time.Minute)
			data.GetCoinTradeData("BTC", db)
		}
	}()

	go func() {
		for {
			ctp := dbfunc.Select(db, "BTC", 5)
			tangent := float64((ctp[0].Bolband-ctp[1].Bolband)/ctp[0].Bolband) + 0.005
			ro := data.GetRecentOrder("BTC")
			currentValue, _ := strconv.ParseUint(ro.Ask.Price, 10, 64)
			if currentValue < (ctp[0].Bolband-5*uint64(ctp[0].Bolbandsd)/2) && tangent > 0 {
				logger.Println("Current Value goes lower than BolBand Low Line!")
				weight := (tangent * 100) * 0.5
				available, _ := strconv.ParseFloat(myAccounts.Krw.Available, 64)
				qty := available * weight / float64(currentValue)
				buyID := myAccounts.BuyCoin("BTC", currentValue, qty)

				time.Sleep(time.Duration(15) * time.Minute)
				var i int
				for _, limitOrder := range myLimitOrders.LimitOrders {
					if limitOrder.OrderId == buyID {
						myAccounts.CancelOrder(buyID, currentValue, qty, "bid")
					} else {
						i++
					}
				}
				if i == len(myLimitOrders.LimitOrders) {
					ro = data.GetRecentOrder("BTC")
					currentValue, _ := strconv.ParseUint(ro.Bid.Price, 10, 64)
					myAccounts.SellCoin("BTC", currentValue, qty)
				}
			} else {
				logger.Println("Current Value is in Bollinger Band")
			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(10) * time.Second)
			logger.Println("Let's Delete unresolved Buy/Sell request.")
			if len(myLimitOrders.LimitOrders) == 0 {
				logger.Println("No Order Exists")
			}
			for _, limitOrder := range myLimitOrders.LimitOrders {
				currentTime := time.Now().Unix()
				timestamp, err := strconv.ParseInt(limitOrder.Timestamp, 10, 64)
				if err != nil {
					logger.Println(err)
				}
				if timestamp < currentTime-3600 {
					logger.Println("Cancelling a Order" + limitOrder.OrderId)
					price, _ := strconv.ParseUint(limitOrder.Price, 10, 64)
					qty, _ := strconv.ParseFloat(limitOrder.Qty, 64)
					myAccounts.CancelOrder(limitOrder.OrderId, price, qty, limitOrder.Type)
				} else {
					logger.Println(limitOrder.OrderId + " will be deleted in " + strconv.Itoa(int(currentTime-timestamp)+3600))
				}
			}
		}
	}()

	for {
		fmt.Scanln()
	}
}
