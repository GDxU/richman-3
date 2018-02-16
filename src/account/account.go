package account

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"logger"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Secret values of accounts
const (
	AccessToken string = "472e5219-1653-4d31-a7fc-28040de08d00"
	SecretKey   string = "c9c24f88-8f78-4e34-ace2-c8bba0e52d51"
)

// URL
const (
	BaseURL string = "https://api.coinone.co.kr"
)

// Parameter is the form of CoinOne request
// related to all services about accounts.
type Parameter struct {
	Access_token string
	Order_id     string
	Price        uint64
	Qty          float64
	Is_ask       int
	Currency     string
	Nonce        uint
}

// MyBalance is balance.
type MyBalance struct {
	Result       string
	ErrorCode    string
	ErrorMessage string
	Btc          struct {
		Available string
		Balance   string
	}
	Ltc struct {
		Available string
		Balance   string
	}
	Eth struct {
		Available string
		Balance   string
	}
	Xrp struct {
		Available string
		Balance   string
	}
	Qtum struct {
		Available string
		Balance   string
	}
	Iota struct {
		Available string
		Balance   string
	}
	Krw struct {
		Available string
		Balance   string
	}
}

// MyLimitOrders is limit orders
type MyLimitOrders struct {
	Result       string
	ErrorCode    string
	ErrorMessage string
	LimitOrders  []struct {
		Index     string
		Timestamp string
		Price     string
		Qty       string
		OrderId   string
		Type      string
	}
}

// CancelRes is the response of Cancelling the Order
type CancelRes struct {
	Result    string
	ErrorCode string
}

// LimitBuySellRes body
type LimitBuySellRes struct {
	Result    string
	ErrorCode string
	OrderId   string
}

// OrderInfoRes body
type OrderInfoRes struct {
	Result    string
	ErrorCode string
	Status    string
	Info      []struct {
		Price     string
		Timestamp string
		Qty       string
		RemainQty string
		Type      string
		Currency  string
		OrderId   string
	}
}

// GetBalance gets a balance of the account.
func GetBalance() *MyBalance {
	logger := logger.GetLogger("[Get Balance]")

	url := BaseURL + "/v2/account/balance/"

	p := Parameter{
		Access_token: AccessToken,
		Nonce:        uint(time.Now().Unix()),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	b := new(MyBalance)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(b)
		if err2 == nil {
			logger.Println("Get Balance Succeeded.")
		} else {
			logger.Println(resp.Body)
			logger.Println(err2)
		}
	} else {
		logger.Println(err)
	}
	return b
}

// BuyCoin sends a request for limit buy
// @return: OrderId or ErrorCode or EmptyString
func (b *MyBalance) BuyCoin(coin string, price uint64, qty float64) string {
	logger := logger.GetLogger("[Buy " + coin + "Coins]")
	url := BaseURL + "/v2/order/limit_buy/"

	p := Parameter{
		Access_token: AccessToken,
		Price:        price,
		Qty:          qty,
		Currency:     coin,
		Nonce:        uint(time.Now().Unix()),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	lbs := new(LimitBuySellRes)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(lbs)
		if err2 == nil {
			if lbs.Result == "success" {
				logger.Println("Request for a Limit Buy Succeeded.")
				return lbs.OrderId
			}
			return lbs.ErrorCode
		}
		logger.Println(err2)
		return ""
	} else {
		logger.Println(err)
		return ""
	}
}

// SellCoin registers a limit sell request
// @return: OrderId or ErrorCode or EmptyString
func (b *MyBalance) SellCoin(coin string, price uint64, qty float64) string {
	logger := logger.GetLogger("[Sell " + coin + " Coins]")
	url := BaseURL + "/v2/order/limit_sell/"

	p := Parameter{
		Access_token: AccessToken,
		Nonce:        uint(time.Now().Unix()),
		Price:        price,
		Qty:          qty,
		Currency:     coin,
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}
	lbs := new(LimitBuySellRes)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(lbs)
		if err2 == nil {
			if lbs.Result == "success" {
				logger.Println("Request for a Limit Sell Succeeded.")
				return lbs.OrderId
			}
			return lbs.ErrorCode
		}
		logger.Println(err2)
		return ""
	} else {
		logger.Println(err)
		return ""
	}
}

// CancelOrder cancels an order, if fails, empty string or errorCode will return.
// Else, canceled orderId returns.
func (b *MyBalance) CancelOrder(id string, price uint64, qty float64, tradeType string) string {
	logger := logger.GetLogger("[Cancel Order]")
	url := BaseURL + "/v2/order/cancel/"

	var isAsk int
	if tradeType == "ask" {
		isAsk = 1
	} else {
		isAsk = 0
	}

	p := Parameter{
		Access_token: AccessToken,
		Order_id:     id,
		Nonce:        uint(time.Now().Unix()),
		Qty:          qty,
		Price:        price,
		Is_ask:       isAsk,
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	c := new(CancelRes)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(c)
		if err2 == nil {
			if c.Result == "success" {
				logger.Println("Cancel an Order Succeeded : " + id)
				logger.Println("Qty : " + strconv.FormatFloat(qty, 'g', 1, 64))
				logger.Println("Price : " + strconv.FormatUint(price, 10))
				return id
			}
			return c.ErrorCode
		}
		logger.Println(err2)
		return ""
	} else {
		logger.Println(err)
		return ""
	}
}

// GetOrderInfo returns an Order Info with given OrderId
func (b *MyBalance) GetOrderInfo(coin, orderId string) *OrderInfoRes {
	logger := logger.GetLogger("[Get Order Info]")
	url := BaseURL + "/v2/order/order_info/"

	p := Parameter{
		Access_token: AccessToken,
		Currency:     coin,
		Order_id:     orderId,
		Nonce:        uint(time.Now().Unix()),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	oir := new(OrderInfoRes)

	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(oir)
		if err2 == nil {
			if oir.Result == "success" {
				logger.Println("Get Order Info Succeeded : " + orderId)
				return oir
			}
			logger.Println(oir)
			return nil
		}
		logger.Println(err2)
		return nil
	} else {
		logger.Println(err)
		return nil
	}
}

// GetLimitOrders returns all un-traded orders.
func (b *MyBalance) GetLimitOrders(coin string) *MyLimitOrders {
	logger := logger.GetLogger("[Get Limit Orders]")
	url := BaseURL + "/v2/order/limit_orders/"

	p := Parameter{
		Access_token: AccessToken,
		Currency:     coin,
		Nonce:        uint(time.Now().Unix()),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	l := new(MyLimitOrders)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(l)
		if err2 == nil {
			if l.Result == "success" {
				logger.Println("Get LimitOrders Succeeded.")
				return l
			}
			return nil
		}
		logger.Println(err2)
		return nil
	} else {
		logger.Println(err)
		return nil
	}
}

// setRequest transforms a golang data into a coinone request form.
func (p *Parameter) setRequest(url string, logger *log.Logger) *http.Request {

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(p)

	// if a go structure has fields which has lower letter in their first char
	// other libraries can't use those data, so cannot encode or transform into a proper form.
	lowerB := []byte(strings.ToLower(b.String()))

	encodedPayload := base64.StdEncoding.EncodeToString(lowerB)

	hash := hmac.New(sha512.New, []byte(strings.ToUpper(SecretKey)))
	hash.Write([]byte(encodedPayload))

	signature := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logger.Println(err)
	}
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-COINONE-PAYLOAD", encodedPayload)
	req.Header.Add("X-COINONE-SIGNATURE", signature)

	return req
}
