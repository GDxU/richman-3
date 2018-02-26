package logics

import (
	"account"
	"data"
	"database/sql"
	"dbfunc"
	"logger"
	"strconv"
	"sync"
	"time"
)

var (
	errorMessage string
)

// LogicA ...
func LogicA(ma *account.MyBalance, mlo *account.MyLimitOrders, mco *account.MyCompleteOrders,
	db *sql.DB, coin string, lock *sync.RWMutex) {
	logger := logger.GetLogger("[Logic A]")
	var (
		sellValue, currentValue uint64
		tangent                 float64
		ctp                     []dbfunc.CoinTradePrice
		ro                      *data.RecentOrderBook
	)

A:
	ctp = dbfunc.Select(db, coin, 5)
	if ctp == nil {
		errorMessage = "CoinTradePrice Fetch Failed."
		goto RETURN
	}
	tangent = float64(int(ctp[4].Bolband)-int(ctp[3].Bolband))/float64(ctp[4].Bolband) + 0.005
	ro = data.GetRecentOrder(coin)
	if ro == nil {
		errorMessage = "GetRecentOrder Failed."
		goto RETURN
	}

	currentValue, _ = strconv.ParseUint(ro.Ask.Price, 10, 64)
	if currentValue < (ctp[0].Bolband-5*uint64(ctp[0].Bolbandsd)/2) && tangent > 0 {
		logger.Info.Println("Current Value goes lower than BolBand Low Line! : " + strconv.Itoa(int(currentValue)) +
			" krw now, " + strconv.Itoa(int((ctp[0].Bolband - 5*uint64(ctp[0].Bolbandsd)/2))) + " krw LowerLine of BolBand")
		weight := tangent * 100
		available, err := strconv.ParseFloat(ma.Krw.Avail, 64)
		if err != nil {
			logger.Warning.Println(err)
			goto A
		}
		qty := available * weight / float64(currentValue)
		buyID := account.BuyCoin(coin, currentValue, qty)
		if len(buyID) <= 5 {
			errorMessage = "Buy Coin Failed, ErrorCode : " + buyID
			goto RETURN
		}

		time.Sleep(time.Duration(15) * time.Minute)

		for _, limitOrder := range mlo.LimitOrders {
			if limitOrder.OrderId == buyID {
				account.CancelOrder(buyID, currentValue, qty, "bid")
				return
			}
		}

		for scount := 0; 0 < 30; scount++ {
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
				if scount == 29 {
					logger.Info.Println("Let's Jonh-Ber")
					return
				}
				time.Sleep(time.Duration(2) * time.Minute)
				continue
			} else {
				break
			}
		}
		for count := 0; count < 10; count++ {
			sellID := account.SellCoin(coin, sellValue, qty)
			if sellID == "" {
				if count == 9 {
					errorMessage = "Sell Coin Failed, SellID : " + sellID
					return
				}
				time.Sleep(time.Duration(1) * time.Second)
				continue
			} else {
				return
			}
		}
	}
	return

RETURN:
	logger.Severe.Println(errorMessage)
	return
}

func LogicB(ma *account.MyBalance, mlo *account.MyLimitOrders, mco *account.MyCompleteOrders,
	db *sql.DB, coin string, lock *sync.RWMutex) {
	logger := logger.GetLogger("[Logic A]")
	var (
		sellValue, currentValue uint64
		weight                  float64
		ctp                     []dbfunc.CoinTradePrice
		ro                      *data.RecentOrderBook
	)
	ctp = dbfunc.Select(db, coin, 5)
	if ctp == nil {
		errorMessage = "CoinTradePrice Fetch Failed."
		goto RETURN
	}

	ro = data.GetRecentOrder(coin)
	if ro == nil {
		errorMessage = "GetRecentOrder Failed."
		goto RETURN
	}
	currentValue, _ = strconv.ParseUint(ro.Ask.Price, 10, 64)

A:
	if ctp[0].LastPrice > ctp[0].FirstPrice && ctp[1].LastPrice > ctp[1].LastPrice &&
		ctp[2].LastPrice > ctp[2].FirstPrice && ctp[3].LastPrice > ctp[3].FirstPrice &&
		ctp[4].LastPrice > ctp[4].FirstPrice && ctp[0].AvgPrice > ctp[2].Bolband && ctp[2].Bolband > ctp[4].AvgPrice {

		logger.Info.Println(coin + " is RISING!")
		weight = 0.5
		available, err := strconv.ParseFloat(ma.Krw.Avail, 64)
		if err != nil {
			logger.Warning.Println(err)
			goto A
		}
		qty := available * weight / float64(currentValue)
		buyID := account.BuyCoin(coin, currentValue, qty)
		if len(buyID) <= 5 {
			errorMessage = "Buy Coin Failed, ErrorCode : " + buyID
			goto RETURN
		}

		time.Sleep(time.Duration(60) * time.Minute)
		for _, limitOrder := range mlo.LimitOrders {
			if limitOrder.OrderId == buyID {
				account.CancelOrder(buyID, currentValue, qty, "bid")
				return
			}
		}
		for scount := 0; 0 < 30; scount++ {
			for {
				ro = data.GetRecentOrder(coin)
				if ro == nil {
					time.Sleep(time.Duration(1) * time.Second)
					continue
				}
				break
			}
			sellValue, _ = strconv.ParseUint(ro.Bid.Price, 10, 64)
			if sellValue < currentValue {
				if scount == 29 {
					logger.Info.Println("Let's Jonh-Ber")
					return
				}
				time.Sleep(time.Duration(2) * time.Minute)
				continue
			}
			break
		}
		for count := 0; count < 10; count++ {
			sellID := account.SellCoin(coin, sellValue, qty)
			if sellID == "" {
				if count == 9 {
					errorMessage = "Sell Coin Failed, SellID : " + sellID
					return
				}
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}
			return
		}
	}
	return

RETURN:
	logger.Severe.Println(errorMessage)
	return
}
