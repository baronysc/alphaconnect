package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"env"
	"fmt"
	"msg"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

// ArmorKey 用來提供給 hmac 轉換，並驗證傳值的正確性
const ArmorKey string = "OQrdcqpv26hBr8ef"

// GenSign 驗證傳值的正確性
func GenSign(appkey string, data string) string {
	key := []byte(appkey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(data))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return sign
}

func main() {

	fmt.Println("Provide alpha get mongo db data")

	env.Start()

	spew.Dump(env.Data)

	msg.SetSubjoinFileName("alpha")
	msg.Start()

	startListen := make(chan bool, 1)

	go func() {
		http.HandleFunc("/", JokeBack)

		startListen <- true
		fmt.Println("start listen")

		http.ListenAndServe(":4321", nil)
	}()

	// wait start listen
	if <-startListen == false {

	}

	fmt.Println("listen ...")

	select {}

}

// serverSayHello 用來回應給連線不帶參數的 client 使用
type _ServerSayHello struct {
	ClientIP string `json:"ClientIP"`
}

// JokeBack 用來給 client 直接讀取判斷 ip 使用
func JokeBack(w http.ResponseWriter, r *http.Request) {

	content := _ServerSayHello{
		ClientIP: r.RemoteAddr,
	}

	b, _ := json.Marshal(content)

	w.Write([]byte(b))

}
