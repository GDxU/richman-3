package dbfunc

import (
	"fmt"
	"database/sql"
	"math"
	_ "github.com/go-sql-driver/mysql"
)

const DBUSER string = "rich"
const DBAUTH string = "rich"
const DBIPADDR string = "localhost"
const DBPORT string = "3306"

type PRICE struct {
	Timestamp uint64
	Qty float64
	AvgPrice uint64
	FirstPrice uint64
	LastPrice uint64
	MaxPrice uint64
	MinPrice uint64
	Bolband uint64
	Bolbandsd uint32
}

func (p *PRICE) Insert(db *sql.DB) {
	//insert 
	stmtIns, err := db.Prepare("Insert into btc10min (timestamp, qty, avgPrice, firstPrice, lastPrice, maxPrice,"+
		" minPrice) VALUES (?, ?, ?, ?, ?, ?, ?)")
	defer stmtIns.Close()

	if err != nil {
		fmt.Println(err)
	} else {
		if _, err2 := stmtIns.Exec(p.Timestamp, p.Qty, p.AvgPrice, p.FirstPrice, p.LastPrice, 
			p.MaxPrice, p.MinPrice); err2 != nil {
				fmt.Println(err2)
		}
	}

	var avgPrices []float64
	var avgPrice float64
	var avg float64
	var avgStdDev float64 

	//select avg of avgPrice for bolban update
	row, err := db.Query("select avg(avgPrice) as avg from btc10min where id >= (select max(id) from btc10min) - 20")
	defer row.Close()
	if err != nil {
		fmt.Println(err)
	} else {
		if row.Next(){
			err2 := row.Scan(&avg)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				fmt.Printf("avg value : %f\n", avg)
			}
		}
	}
	//select avgPrices for bolban update
	rows, err := db.Query("select avgPrice from btc10min where id >= (select max(id) from btc10min) - 20")
	defer rows.Close()
	count := 0
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next(){
			err2 := rows.Scan(&avgPrice)
			if err2 != nil {
				fmt.Println(err2)
			} else {
				fmt.Printf("avg values comming : %f\n", avgPrice)
				avgPrices = append(avgPrices, avgPrice)
			}
			count = count + 1
		}
		//calc stdDev
		sum := 0.0
		for _, avgPrice := range avgPrices{
			sum = sum + math.Pow((avgPrice - avg), 2)
		}
		avgStdDev = math.Sqrt(sum / float64(len(avgPrices)))
	}
 
	//update for stddev/mean
	stmtUpdate, err := db.Prepare("Update btc10min set bolband = ?, bolbandsd = ? "+
		"where timestamp = ?")
	defer stmtUpdate.Close()

	if err != nil {
		fmt.Println(err)
	} else {
		if _, err2 := stmtUpdate.Exec(uint64(avg), uint64(avgStdDev), p.Timestamp); err2 != nil {
			fmt.Println(err2)
		}
	}
}