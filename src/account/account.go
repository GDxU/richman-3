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
	"strings"
	"time"
)

const AccessToken string = "472e5219-1653-4d31-a7fc-28040de08d00"
const SecretKey string = "c9c24f88-8f78-4e34-ace2-c8bba0e52d51"
const BaseURL string = "https://api.coinone.co.kr"

type Parameter struct {
	Access_token string
	Order_id     string
	Price        uint64
	Qty          float64
	Is_ask       int
	Currenct     string
	Nonce        uint
}

type Balance struct {
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

func GetBalance() *Balance {
	logger := logger.GetLogger("[Get Balance]")

	url := BaseURL + "/v2/account/balance/"

	p := Parameter{
		Access_token: AccessToken,
		Nonce:        uint(time.Now().Unix()),
	}

	req := p.setRequest(url, logger)

	client := &http.Client{}
	b := new(Balance)
	if resp, err := client.Do(req); err == nil {
		err2 := json.NewDecoder(resp.Body).Decode(b)
		if err2 == nil {
			logger.Println("Get Balance Succeeded.")
		} else {
			logger.Println(err2)
		}
	} else {
		logger.Println(err)
	}
	return b
}

// func (b *Balance) BuyCoin(coin, amount string) bool {
// 	logger := logger.GetLogger("[Buy Coins]")
// 	url := BaseURL +
// }

// func (b *Balance) CancelOrder(id string) bool {
// 	logger.GetLogger("[Cancel Order]")
// }

// func (b *Balance) SellCoin(coin, amount string) bool {

// }

func (p *Parameter) setRequest(url string, logger *log.Logger) *http.Request {

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(p)
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
