package main

import (
	"account"
	"data"
	"database/sql"
	"dbfunc"
	"fmt"
	"logger"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	logger.Now = time.Now().Format(time.RFC822)
	logger := logger.GetLogger("[Let's get Rich]")
	logger.Println("Let's Get Start!")

	db := getDbConn()

	go func() {
		data.GetCoinData("BTC", 10, db)
	}()

	go func() {
		for {
			account.GetBalance()
			time.Sleep(time.Duration(2) * time.Second)
		}
	}()

	fmt.Scanln()
}

func getDbConn() *sql.DB {
	logger := logger.GetLogger("[DBConnection]")
	db, err := sql.Open("mysql", dbfunc.DBUSER+":"+dbfunc.DBAUTH+"@tcp("+dbfunc.DBIPADDR+":"+dbfunc.DBPORT+")/btc")
	err2 := db.Ping()
	if err != nil {
		panic(err.Error)
	} else if err2 != nil {
		fmt.Println(err2.Error())
		panic(err2.Error)
	} else {
		logger.Println("DB Connected.")
	}
	return db
}
