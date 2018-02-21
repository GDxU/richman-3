package main

import (
	"account"
	"bufio"
	"data"
	"dbfunc"
	"fmt"
	"logger"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// Arguments
	args := os.Args[1:]
	reader := bufio.NewReader(os.Stdin)
	var coin string
	if len(args) == 0 {
		fmt.Print("Enter a Coin Name: ")
		text, _ := reader.ReadString('\n')
		coin = strings.Trim(text, "\n")
	} else {
		coin = strings.ToUpper(args[0])
	}

	//Logger Setting
	logger.Now = time.Now().Format(time.RFC822)
	logger.Coin = coin
	mainLogger := logger.GetLogger("[Let's get Rich]")
	mainLogger.Info.Println("Let's Get Start!")

	// TODO
	// How It works?
	defer func() {
		if r := recover(); r != nil {
			mainLogger.Severe.Println(r)
		}
	}()

	var myAccounts *account.MyBalance
	var myLimitOrders *account.MyLimitOrders
	db := dbfunc.GetDbConn(coin)

	// get Account Info every 10 seconds
	go func() {
		for {
			myAccounts = account.GetBalance()
			myLimitOrders = myAccounts.GetLimitOrders(coin)
			if myAccounts == nil || myLimitOrders == nil {
				time.Sleep(time.Duration(1) * time.Second)
				if myAccounts == nil {
					mainLogger.Warning.Println("Get MyAccounts Failed.")
				} else {
					mainLogger.Warning.Println("Get Limit Orders Failed.")
				}
				continue
			}
			time.Sleep(time.Duration(10) * time.Second)
		}
	}()

	// get BTC Trade data every 10 minutes.
	godtCount := 0
	for {
		godt := data.GetOneDayTradeData(coin, db)
		if godt == "" {
			godtCount++
			if godtCount > 5 {
				panic("Get one day trade data failed. Somethings Wrong")
			}
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}
		godtCount = 0
		break
	}

	gctdCount := 0
	go func() {
		for {
			time.Sleep(time.Duration(10) * time.Minute)
			gctd := data.GetCoinTradeData(coin, db)
			if gctd == "" {
				gctdCount++
				if gctdCount > 5 {
					panic("Get coin trade data failed.")
				}
				continue
			}
			gctdCount = 0
		}
	}()

	//Logic A
	var sellValue, currentValue uint64
	go func() {
		logger := logger.GetLogger("[Logic A]")
		for {
			time.Sleep(time.Duration(5) * time.Second)

			ctp := dbfunc.Select(db, coin, 5)
			tangent := float64(int(ctp[4].Bolband)-int(ctp[3].Bolband))/float64(ctp[4].Bolband) + 0.005
			ro := data.GetRecentOrder(coin)
			if ro == nil {
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}
			currentValue, _ = strconv.ParseUint(ro.Ask.Price, 10, 64)

			if currentValue < (ctp[0].Bolband-5*uint64(ctp[0].Bolbandsd)/2) && tangent > 0 {
				logger.Info.Println("Current Value goes lower than BolBand Low Line! : " + strconv.Itoa(int(currentValue)) +
					" krw now, " + strconv.Itoa(int((ctp[0].Bolband - 5*uint64(ctp[0].Bolbandsd)/2))) + " krw LowerLine of BolBand")
				weight := tangent * 100
				available, err := strconv.ParseFloat(myAccounts.Krw.Avail, 64)
				if err != nil {
					logger.Warning.Println(err)
					continue
				}
				qty := available * weight / float64(currentValue)
				buyID := myAccounts.BuyCoin(coin, currentValue, qty)
				if buyID == "" {
					time.Sleep(time.Duration(1) * time.Second)
					continue
				}
				if len(buyID) <= 5 {
					logger.Warning.Println("ErrorCode : " + buyID)
					continue
				}

				for {
					temp1 := 0
					time.Sleep(time.Duration(15) * time.Minute)
					for {
						ro = data.GetRecentOrder(coin)
						if ro == nil {
							time.Sleep(time.Duration(1) * time.Second)
							continue
						} else {
							break
						}
					}
					sellValue, _ = strconv.ParseUint(ro.Bid.Price, 10, 64)
					if sellValue < currentValue {
						temp1++
						if temp1 < 5 {
							continue
						} else {
							break
						}
					}
				}

				var i int
				for _, limitOrder := range myLimitOrders.LimitOrders {
					if limitOrder.OrderId == buyID {
						myAccounts.CancelOrder(buyID, currentValue, qty, "bid")
					} else {
						i++
					}
				}
				if i == len(myLimitOrders.LimitOrders) {

					for {
						sellID := myAccounts.SellCoin(coin, sellValue, qty)
						if sellID == "" {
							time.Sleep(time.Duration(1) * time.Second)
							continue
						} else {
							break
						}
					}
				}
			} else {
				logger.Info.Println("Current Value is in Bollinger Band : " + strconv.Itoa(int(currentValue)) +
					" krw now, " + strconv.Itoa(int((ctp[0].Bolband - 5*uint64(ctp[0].Bolbandsd)/2))) + " krw LowerLine of BolBand")
			}
		}
	}()

	// remove unresolved Buy/Sell request every 10 min
	go func() {
		logger := logger.GetLogger("[Remove Unresolved Buy/Sell Request]")
	Loop1:
		for {
			time.Sleep(time.Duration(10) * time.Second)
			logger.Info.Println("Let's Delete unresolved Buy/Sell request.")
			if len((*myLimitOrders).LimitOrders) == 0 {
				logger.Info.Println("No Order Exists")
				continue Loop1
			}
		Loop2:
			for _, limitOrder := range myLimitOrders.LimitOrders {
				currentTime := time.Now().Unix()
				timestamp, err := strconv.ParseInt(limitOrder.Timestamp, 10, 64)
				if err != nil {
					logger.Warning.Println(err)
					continue Loop2
				}
				if timestamp < currentTime-3600 {
					logger.Info.Println("Cancelling an Order" + limitOrder.OrderId)
					price, _ := strconv.ParseUint(limitOrder.Price, 10, 64)
					qty, _ := strconv.ParseFloat(limitOrder.Qty, 64)

					coCount := 0
				Loop3:
					for {
						cancelID := myAccounts.CancelOrder(limitOrder.OrderId, price, qty, limitOrder.Type)
						if cancelID == "" {
							time.Sleep(time.Duration(1) * time.Second)
							coCount++
							if coCount > 5 {
								logger.Severe.Println("Canceling Order Failed. " + limitOrder.OrderId)
								continue Loop2
							}
							continue Loop3
						} else {
							break Loop3
						}
					}
				} else {
					logger.Info.Println(limitOrder.OrderId + " will be deleted in " + strconv.Itoa(int(currentTime-timestamp)+3600))
				}
			}
		}
	}()

	var mco *account.MyCompleteOrders
	// Logging Complete Trade
	go func() {
		logger := logger.GetLogger("[Complete Trade]")
		for {
			mco = account.GetCompleteOrder(coin)
			if mco == nil {
				time.Sleep(time.Duration(1) * time.Second)
				continue
			} else {
				break
			}
		}
		noco := len(mco.CompleteOrders)

		for {
			time.Sleep(time.Duration(10) * time.Second)
			mco = account.GetCompleteOrder(coin)
			if mco == nil {
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}
			if len(mco.CompleteOrders) > noco {
				for i := 0; i < len(mco.CompleteOrders)-noco; i++ {
					if mco.CompleteOrders[i].Type == "ask" {
						logger.Info.Println("Sell " + coin + " Succeeded.")
					} else {
						logger.Info.Println("Buy " + coin + " Succeeded.")
					}
					logger.Info.Println(mco.CompleteOrders[i].Price + " KRW, " + mco.CompleteOrders[i].Qty + coin)
				}
			}
			noco = len(mco.CompleteOrders)
		}
	}()

	for {
		fmt.Scanln()
	}
}
