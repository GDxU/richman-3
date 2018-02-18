package main

import "fmt"

func main() {
	a := "whojes"
	if a == "whojes" {
		fmt.Println("whojes")
	} else {
		fmt.Println("fuck")
	}
}

// mlo := myAccounts.GetLimitOrders("BTC")
// fmt.Println(mlo)

// fmt.Println(myAccounts.BuyCoin("BTC", 100000, 1.0))

// mlo = myAccounts.GetLimitOrders("BTC")
// fmt.Println(mlo)

// limitOrder := mlo.LimitOrders[0]
// p, _ := strconv.Atoi(limitOrder.Price)
// q, _ := strconv.ParseFloat(limitOrder.Qty, 64)

// s := myAccounts.CancelOrder(limitOrder.OrderId, uint64(p), q, limitOrder.Type)
// fmt.Println(s)

// mlo = myAccounts.GetLimitOrders("BTC")
// fmt.Println(mlo)
