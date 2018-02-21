package dbfunc

import (
	"database/sql"
	"fmt"
	"logger"
	"math"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Constants for DB Connection
const (
	DBUSER   string = "rich"
	DBAUTH   string = "rich"
	DBIPADDR string = "localhost"
	DBPORT   string = "3306"
)

// CoinTradePrice is the Dataform for DB Insertion
// TABLE NAME FORM: btc10min
type CoinTradePrice struct {
	ID         uint64
	Timestamp1 uint64
	Timestamp2 uint64
	Qty        float64
	AvgPrice   uint64
	FirstPrice uint64
	LastPrice  uint64
	MaxPrice   uint64
	MinPrice   uint64
	Bolband    uint64
	Bolbandsd  uint32
}

// Insert inserts CoinTradeData into Database
// It does BOLLENGER BAND calculation also.
func (p *CoinTradePrice) Insert(db *sql.DB, coin string) {

	coin = strings.ToLower(coin)
	//insert
	stmtIns, err := db.Prepare("Insert into " + coin + "10min (timestamp1, timestamp2, qty, avgPrice, firstPrice, lastPrice, maxPrice," +
		" minPrice) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	defer stmtIns.Close()

	if err != nil {
		fmt.Println(err)
	} else {
		if _, err2 := stmtIns.Exec(p.Timestamp1, p.Timestamp2, p.Qty, p.AvgPrice, p.FirstPrice, p.LastPrice,
			p.MaxPrice, p.MinPrice); err2 != nil {
			fmt.Println(err2)
		}
	}

	var avgPrices []float64
	var avgPrice float64
	var avg float64
	var avgStdDev float64

	//select avg of avgPrice for bolban update
	row, err := db.Query("select avg(avgPrice) as avg from " + coin + "10min where id >= (select max(id) from " + coin + "10min) - 20")
	defer row.Close()
	if err != nil {
		fmt.Println(err)
	} else {
		if row.Next() {
			err2 := row.Scan(&avg)
			if err2 != nil {
				fmt.Println(err2)
			} else {
			}
		}
	}
	//select avgPrices for bolban update
	rows, err := db.Query("select avgPrice from " + coin + "10min where id >= (select max(id) from " + coin + "10min) - 20")
	defer rows.Close()
	count := 0
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next() {
			err2 := rows.Scan(&avgPrice)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				avgPrices = append(avgPrices, avgPrice)
			}
			count = count + 1
		}
		//calc stdDev
		sum := 0.0
		for _, avgPrice := range avgPrices {
			sum = sum + math.Pow((avgPrice-avg), 2)
		}
		avgStdDev = math.Sqrt(sum / float64(len(avgPrices)))
	}

	//update for stddev/mean
	stmtUpdate, err := db.Prepare("Update " + coin + "10min set bolband = ?, bolbandsd = ? " +
		"where timestamp1 = ?")
	defer stmtUpdate.Close()

	if err != nil {
		fmt.Println(err)
	} else {
		if _, err2 := stmtUpdate.Exec(uint64(avg), uint64(avgStdDev), p.Timestamp1); err2 != nil {
			fmt.Println(err2)
		}
	}

	return
}

// Select fetches 20 Rows of Coin Data
// @param: coin
func Select(db *sql.DB, coin string, count int) []CoinTradePrice {
	logger := logger.GetLogger("[Select From Database]")
	coin = strings.ToLower(coin)
	//select
	rows, err := db.Query("Select id, timestamp1, timestamp2, avgPrice, bolband, bolbandsd, firstPrice, lastPrice," +
		" maxPrice, minPrice, qty from " + coin + "10min where id >= (select max(id) from " + coin + "10min) - " + strconv.Itoa(count-1))
	if err != nil {
		logger.Severe.Println(err)
		return nil
	}
	defer rows.Close()

	arrCtp := []CoinTradePrice{}

	for rows.Next() {
		var ctp CoinTradePrice
		err2 := rows.Scan(&ctp.ID, &ctp.Timestamp1, &ctp.Timestamp2, &ctp.AvgPrice, &ctp.Bolband, &ctp.Bolbandsd,
			&ctp.FirstPrice, &ctp.LastPrice, &ctp.MaxPrice, &ctp.MinPrice, &ctp.Qty)
		if err2 != nil {
			logger.Severe.Println(err2)
			return nil
		}
		arrCtp = append(arrCtp, ctp)
	}

	return arrCtp
}

// GetDbConn returns a connectable DB Connection
// Parameters for connection are constants of the package
func GetDbConn(coin string) *sql.DB {
	logger := logger.GetLogger("[DBConnection]")
	coin = strings.ToLower(coin)
	db, err := sql.Open("mysql", DBUSER+":"+DBAUTH+"@tcp("+DBIPADDR+":"+DBPORT+")/"+coin)
	err2 := db.Ping()
	if err != nil {
		logger.Severe.Println(err)
		return nil
	} else if err2 != nil {
		logger.Severe.Println(err2)
		return nil
	}
	logger.Info.Println("DB Connected.")
	return db
}
