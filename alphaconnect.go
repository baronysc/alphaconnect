package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"env"
	"fmt"
	"io/ioutil"
	"mgodb"
	"msg"
	"net/http"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

type _RecvData struct {
	Sign string      `json:"sign"`
	Data interface{} `json:"data"`
}

func main() {

	fmt.Println("Provide alpha get mongo db data")

	env.Start()

	spew.Dump(env.Data)

	msg.SetSubjoinFileName("alpha") // 附加額外的 log 檔案名稱，若不執行則為空字串
	msg.Start()

	if mgodb.Start() != nil {
		color.Set(color.FgRed)
		fmt.Println("ERROR: mongodb run fail")
		color.Unset()
		msg.Log("ERROR: mongodb run fail")
		return
	}

	startListen := make(chan bool, 1)

	go func() {
		http.HandleFunc("/", JokeBack)
		http.HandleFunc("/Registered", Registered)
		http.HandleFunc("/Playerinfo", Playerinfo)
		http.HandleFunc("/Rankinfo", Rankinfo)

		startListen <- true
		fmt.Println("start listen")

		http.ListenAndServe("0.0.0.0:8082", nil)
	}()

	// wait start listen
	if <-startListen == false {

	}

	fmt.Println("listen ...")

	select {}

}

// ReturnData 用來準備回傳資料使用
type ReturnData struct {
	State   int         `json:"state"`
	Text    string      `json:"text"`
	Error   interface{} `json:"error, omitempty"`
	Content interface{} `json:"content, omitempty"`
}

// AlphaKey 用來提供給 hmac 轉換，並驗證傳值的正確性
const AlphaKey string = "OQrdcqpv26hBr8ef"

// GenSign 驗證傳值的正確性, 之前 server 的寫法
func GenSign(appkey string, data string) string {
	key := []byte(appkey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(data))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return sign
}

// signature 簽章判斷
func (r *ReturnData) signature(_Sign string, _Data string) bool {

	key := []byte(AlphaKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(_Data))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if strings.Compare(sign, _Sign) != 0 {
		r.State = -1
		return false
	}

	return true
}

func (r *ReturnData) jsonContent(_Content interface{}) []byte {

	rtnData, err := json.Marshal(_Content)

	if err != nil {
		// 列出失敗的資訊
		v := reflect.ValueOf(_Content)
		t := v.Type()
		msg.Log("ERROR: struct name <<", t.Name(), ">> json marshal fail.")
		for i := 0; i < v.NumField(); i++ {
			msg.Log(t.Field(i).Name, v.Field(i).Type(), v.Field(i).Interface())
		}
		rtnData = []byte("ERROR:result data json marshal fail")
	}

	return rtnData
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

type _IDinfo struct {
	IgsID   int    `json:"igsid`
	AlphaID string `json:"alphaid"`
}

type _RecvPlayerData struct {
	Sign string  `json:"sign"`
	Data _IDinfo `json:"data"`
}

// Registered 用來給 client 註冊奧飛通行證資料
func Registered(w http.ResponseWriter, r *http.Request) {

	// 未做認證

	result := mgodb.AlphaData.Registered(3, "3")

	resultString := "Registered:" + result.Error()

	b, _ := json.Marshal(resultString)

	w.Write([]byte(b))

}

// Playerinfo 用來給 client 直接讀玩家資料使用
func Playerinfo(w http.ResponseWriter, r *http.Request) {

	// 未做認證

	info, err := mgodb.AlphaData.Playerinfo(3, "3")

	if err != nil {
		b, _ := json.Marshal("Playerinfo:not found")
		w.Write([]byte(b))
		return
	}

	b, _ := json.Marshal(info)

	w.Write([]byte(b))

}

type _Rankinfo struct {
	Item int `json:"item`
}

type _RecvRankData struct {
	Sign string    `json:"sign"`
	Data _Rankinfo `json:"data"`
}

// Rankinfo 用來給 client 直接讀取排名資料使用
func Rankinfo(w http.ResponseWriter, r *http.Request) {

	body, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvRankData{}
	err := json.Unmarshal(body, &recvData)
	if err != nil {
		fmt.Println(err)
		b, _ := json.Marshal("Rankinfo:json unmarshal fail")
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	returnData := ReturnData{}
	signData, _ := json.Marshal(recvData.Data)
	if returnData.signature(recvData.Sign, string(signData)) == false {
		fmt.Println(err)
		b, _ := json.Marshal("Rankinfo:sign fail")
		w.Write([]byte(b))
		return
	}

	rank, err := mgodb.AlphaData.Rankinfo(recvData.Data.Item)

	if err != nil {
		b, _ := json.Marshal("Rankinfo:not found")
		w.Write([]byte(b))
		return
	}

	b, _ := json.Marshal(rank)

	w.Write([]byte(b))

}
