package main

import (
	"account"
	"bufio"
	"data"
	"database/sql"
	"dbfunc"
	"fmt"
	"logger"
	"logics"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ma         *account.MyBalance
	mlo        *account.MyLimitOrders
	mco        *account.MyCompleteOrders
	lock       *sync.RWMutex
	coin       string
	noco       int
	db         *sql.DB
	errrrr     string
	mainLogger *logger.Loggers
	errchan    chan interface{}
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			mainLogger.Severe.Println(r)
			mainLogger.Severe.Println("=> Unexpected Error.")
		}
	}()
	// Arguments
	args := os.Args[1:]
	reader := bufio.NewReader(os.Stdin)
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
	mainLogger = logger.GetLogger("[Let's get Rich]")
	mainLogger.Info.Println("Let's Get Start!")

	// Base Setting
	errchan = make(chan interface{})
	lock = new(sync.RWMutex)
	db = dbfunc.GetDbConn(coin)
	mco = account.GetCompleteOrder(coin)
	mlo = account.GetLimitOrders(coin)
	ma = account.GetBalance()
	if mco == nil {
		errrrr = "Get Complete Order Failed."
	} else if mlo == nil {
		errrrr = "Get Limit Order Failed."
	} else if ma == nil {
		errrrr = "Get Account Balance Failed."
	} else if db == nil {
		errrrr = "Get DB Connection Failed."
	}
	// get BTC Trade data during past 24 hours
	godt := data.GetOneDayTradeData(coin, db)
	if godt == "" {
		errrrr = "Get one day trade data failed. Somethings Wrong"
	}
	if errrrr != "" {
		panic(errrrr)
	}

	////////////////////////////////////////////////////////////////
	// get Account Info every 10 seconds
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errchan <- r
			}
		}()
		noco = len(mco.CompleteOrders)
		for {
			time.Sleep(time.Duration(10) * time.Second)
			maTemp := account.GetBalance()
			time.Sleep(time.Duration(100) * time.Millisecond)
			mloTemp := account.GetLimitOrders(coin)
			time.Sleep(time.Duration(100) * time.Millisecond)
			mcoTemp := account.GetCompleteOrder(coin)
			if ma == nil || mlo == nil || mcoTemp == nil {
				time.Sleep(time.Duration(1) * time.Second)
				if ma == nil {
					mainLogger.Warning.Println("Get MyAccounts Failed.")
				} else if mlo == nil {
					mainLogger.Warning.Println("Get Limit Orders Failed.")
				} else {
					mainLogger.Warning.Println("Get Complete Order Failed.")
				}
				continue
			}
			lock.Lock()
			*ma = *maTemp
			*mlo = *mloTemp
			*mco = *mcoTemp
			lock.Unlock()

			if noco < len(mco.CompleteOrders) {
				logCompletedTrades(noco)
			}
			noco = len(mco.CompleteOrders)
		}
	}()

	// get BTC Trade data every 10 minutes && logic B
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errchan <- r
			}
		}()
		for {
			time.Sleep(time.Duration(10) * time.Minute)
			getTradeData()
			logics.LogicB(ma, mlo, mco, db, coin, lock)
		}
	}()

	// remove unresolved Buy/Sell request every 60 sec
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errchan <- r
			}
		}()
		removeUnresolvedOrders(60)
	}()

	//Logic A
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errchan <- r
			}
		}()
		for {
			time.Sleep(time.Duration(5) * time.Second)
			logics.LogicA(ma, mlo, mco, db, coin, lock)
		}
	}()

	// how to maintain the program not terminated?
	mainLogger.Severe.Println(<-errchan)
}

func removeUnresolvedOrders(duration int) {
	logger := logger.GetLogger("[Remove Unresolved Buy/Sell Request]")
	for {
		time.Sleep(time.Duration(duration) * time.Second)
		if len(mlo.LimitOrders) == 0 {
			continue
		}
	Loop:
		for _, limitOrder := range mlo.LimitOrders {
			currentTime := time.Now().Unix()
			timestamp, err := strconv.ParseInt(limitOrder.Timestamp, 10, 64)
			if err != nil {
				logger.Warning.Println(err)
				continue Loop
			}
			if timestamp < currentTime-3600 {
				logger.Info.Println("Cancelling an Order" + limitOrder.OrderId)
				price, _ := strconv.ParseUint(limitOrder.Price, 10, 64)
				qty, _ := strconv.ParseFloat(limitOrder.Qty, 64)

			Loop2:
				for coCount := 0; coCount < 5; coCount++ {
					cancelID := account.CancelOrder(limitOrder.OrderId, price, qty, limitOrder.Type)
					if cancelID == "" {
						time.Sleep(time.Duration(1) * time.Second)
						if coCount == 4 {
							logger.Severe.Println("Canceling Order Failed. " + limitOrder.OrderId)
							continue Loop
						}
						continue Loop2
					} else {
						break Loop2
					}
				}
			}
		}
	}
}

func logCompletedTrades(noco int) {
	logger := logger.GetLogger("[Complete Trade]")
	for i := 0; i < len(mco.CompleteOrders)-noco; i++ {
		if mco.CompleteOrders[i].Type == "ask" {
			logger.Info.Println("Sell " + coin + " Succeeded.")
		} else {
			logger.Info.Println("Buy " + coin + " Succeeded.")
		}
		logger.Info.Println(mco.CompleteOrders[i].Price + " KRW, " + mco.CompleteOrders[i].Qty + coin)
	}
}

func getTradeData() {
	count := 0
retry:
	gctd := data.GetCoinTradeData(coin, db)
	if gctd == "" {
		mainLogger.Warning.Println("Get coin trade data failed.")
		time.Sleep(time.Duration(30) * time.Second)
		count++
		if count == 5 {
			panic("Cannot get Coin Trade Data.")
		}
		goto retry
	}
	return
}
