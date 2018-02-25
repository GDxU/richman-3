package account

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"logger"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	Nonce        uint64
}

// MyBalance is balance.
type MyBalance struct {
	Result       string
	ErrorCode    string
	ErrorMessage string
	Btc          struct {
		Avail   string
		Balance string
	}
	Ltc struct {
		Avail   string
		Balance string
	}
	Eth struct {
		Avail   string
		Balance string
	}
	Xrp struct {
		Avail   string
		Balance string
	}
	Qtum struct {
		Avail   string
		Balance string
	}
	Iota struct {
		Avail   string
		Balance string
	}
	Krw struct {
		Avail   string
		Balance string
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

type MyCompleteOrders struct {
	Result         string
	ErrorCode      string
	CompleteOrders []struct {
		Timestamp string
		Price     string
		Qty       string
		Type      string
		OrderID   string
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
		Nonce:        uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	b := new(MyBalance)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(b)
		if err2 == nil {
			if b.Result == "success" {
				return b
			} else if b.ErrorCode == "131" {
				time.Sleep(time.Duration(1) * time.Second)
				return GetBalance()
			} else {
				logger.Warning.Println("[Get Balance Failed " + b.Result + "]")
				return nil
			}
		} else {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			logger.Warning.Println(err2)
			logger.Warning.Println(buf.String())
			return nil
		}
	} else {
		logger.Severe.Println(err)
		return nil
	}
}

// BuyCoin sends a request for limit buy
// @return: OrderId or ErrorCode or EmptyString
func BuyCoin(coin string, price uint64, qty float64) string {
	logger := logger.GetLogger("[Buy " + coin + "Coins]")
	url := BaseURL + "/v2/order/limit_buy/"

	qty = toFixed(qty, 4)

	p := Parameter{
		Access_token: AccessToken,
		Price:        price,
		Qty:          qty,
		Currency:     coin,
		Nonce:        uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	lbs := new(LimitBuySellRes)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(lbs)
		if err2 == nil {
			if lbs.Result == "success" {
				logger.Info.Println("Request for a Limit Buy Succeeded.")
				return lbs.OrderId
			} else if lbs.ErrorCode == "131" {
				time.Sleep(time.Duration(1) * time.Second)
				return BuyCoin(coin, price, qty)
			}
			logger.Warning.Println(lbs.ErrorCode)
			return lbs.ErrorCode
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		logger.Warning.Println(err2)
		logger.Warning.Println(buf.String())
		return ""
	} else {
		logger.Severe.Println(err)
		return ""
	}
}

// SellCoin registers a limit sell request
// @return: OrderId or ErrorCode or EmptyString
func SellCoin(coin string, price uint64, qty float64) string {
	logger := logger.GetLogger("[Sell " + coin + " Coins]")
	url := BaseURL + "/v2/order/limit_sell/"

	qty = toFixed(qty, 4)

	p := Parameter{
		Access_token: AccessToken,
		Nonce:        uint64(time.Now().UnixNano() / int64(time.Millisecond)),
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
				logger.Info.Println("Request for a Limit Sell Succeeded.")
				return lbs.OrderId
			} else if lbs.ErrorCode == "131" {
				time.Sleep(time.Duration(1) * time.Second)
				return SellCoin(coin, price, qty)
			}
			logger.Warning.Println(lbs.ErrorCode)
			return lbs.ErrorCode
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		logger.Warning.Println(err2)
		logger.Warning.Println(buf.String())
		return ""
	} else {
		logger.Severe.Println(err)
		return ""
	}
}

// CancelOrder cancels an order, if fails, empty string or errorCode will return.
// Else, canceled orderId returns.
func CancelOrder(id string, price uint64, qty float64, tradeType string) string {
	logger := logger.GetLogger("[Cancel Order]")
	url := BaseURL + "/v2/order/cancel/"

	var isAsk int
	if tradeType == "ask" {
		isAsk = 1
	} else {
		isAsk = 0
	}

	qty = toFixed(qty, 4)

	p := Parameter{
		Access_token: AccessToken,
		Order_id:     id,
		Nonce:        uint64(time.Now().UnixNano() / int64(time.Millisecond)),
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
				logger.Info.Println("Cancel an Order Succeeded : " + id)
				logger.Info.Println("Qty : " + strconv.FormatFloat(qty, 'g', 1, 64))
				logger.Info.Println("Price : " + strconv.FormatUint(price, 10))
				return id
			} else if c.ErrorCode == "131" {
				time.Sleep(time.Duration(1) * time.Second)
				return CancelOrder(id, price, qty, tradeType)
			}
			return c.ErrorCode
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		logger.Warning.Println(err2)
		logger.Warning.Println(buf.String())
		return ""
	} else {
		logger.Severe.Println(err)
		return ""
	}
}

// GetOrderInfo returns an Order Info with given OrderId
func GetOrderInfo(coin, orderID string) *OrderInfoRes {
	logger := logger.GetLogger("[Get Order Info]")
	url := BaseURL + "/v2/order/order_info/"

	p := Parameter{
		Access_token: AccessToken,
		Currency:     coin,
		Order_id:     orderID,
		Nonce:        uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	oir := new(OrderInfoRes)

	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(oir)
		if err2 == nil {
			if oir.Result == "success" {
				logger.Info.Println("Get Order Info Succeeded : " + orderID)
				return oir
			} else if oir.ErrorCode == "131" {
				time.Sleep(time.Duration(1) * time.Second)
				return GetOrderInfo(coin, orderID)
			}
			logger.Warning.Println(oir.Result)
			return nil
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		logger.Warning.Println(err2)
		logger.Warning.Println(buf.String())
		return nil
	} else {
		logger.Warning.Println(err)
		return nil
	}
}

// GetLimitOrders returns all un-traded orders.
func GetLimitOrders(coin string) *MyLimitOrders {
	logger := logger.GetLogger("[Get Limit Orders]")
	url := BaseURL + "/v2/order/limit_orders/"

	p := Parameter{
		Access_token: AccessToken,
		Currency:     coin,
		Nonce:        uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	l := new(MyLimitOrders)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(l)
		if err2 == nil {
			if l.Result == "success" {
				return l
			} else if l.ErrorCode == "131" {
				time.Sleep(time.Duration(1) * time.Second)
				return GetLimitOrders(coin)
			}
			logger.Warning.Println(l)
			return nil
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		logger.Warning.Println(err2)
		logger.Warning.Println(buf.String())
		return nil
	} else {
		logger.Severe.Println(err)
		return nil
	}
}

// GetCompleteOrder returns complete Orders.
func GetCompleteOrder(coin string) *MyCompleteOrders {
	logger := logger.GetLogger("[Get Complete Orders]")
	url := BaseURL + "/v2/order/complete_orders/"
	p := Parameter{
		Access_token: AccessToken,
		Currency:     coin,
		Nonce:        uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	req := p.setRequest(url, logger)
	client := &http.Client{}

	mco := new(MyCompleteOrders)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(mco)
		if err2 == nil {
			if mco.Result == "success" {
				return mco
			} else if mco.ErrorCode == "131" {
				time.Sleep(time.Duration(1) * time.Second)
				return GetCompleteOrder(coin)
			}
			logger.Warning.Println(mco.ErrorCode)
			return nil
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		logger.Warning.Println(err2)
		logger.Warning.Println(buf.String())
		return nil
	} else {
		logger.Severe.Println(err)
		return nil
	}
}

// setRequest transforms a golang data into a coinone request form.
func (p *Parameter) setRequest(url string, logger *logger.Loggers) *http.Request {

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
		logger.Severe.Println(err)
	}
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-COINONE-PAYLOAD", encodedPayload)
	req.Header.Add("X-COINONE-SIGNATURE", signature)

	return req
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
