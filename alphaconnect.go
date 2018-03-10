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
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

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
		http.HandleFunc("/Playerinfo_Alpha", PlayerinfoAlpha)
		http.HandleFunc("/Playerinfo_Armor", PlayerinfoArmor)
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
	State        int    `json:"state"`
	Message      string `json:"message"`
	JSONString   string `json:"jsonstring"`
	Base64String string `json:"base64string"`
}

// AlphaKey 用來提供給 hmac 轉換，並驗證傳值的正確性
const AlphaKey string = "OQrdcqpv26hBr8ef"

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

func (r *ReturnData) data2JSONString(_Content interface{}) {

	byteArray, err := json.Marshal(_Content)
	if err != nil {
		// 列出失敗的資訊
		v := reflect.ValueOf(_Content)
		t := v.Type()
		msg.Log("ERROR: struct name <<", t.Name(), ">> json marshal fail.")
		for i := 0; i < v.NumField(); i++ {
			msg.Log(t.Field(i).Name, v.Field(i).Type(), v.Field(i).Interface())
		}
		r.Message = "ERROR:failed to convert data to base64"
	}

	r.JSONString = string(byteArray)
}
func (r *ReturnData) data2Base64(_Content interface{}) {

	byteArray, err := json.Marshal(_Content)

	if err != nil {
		// 列出失敗的資訊
		v := reflect.ValueOf(_Content)
		t := v.Type()
		msg.Log("ERROR: struct name <<", t.Name(), ">> json marshal fail.")
		for i := 0; i < v.NumField(); i++ {
			msg.Log(t.Field(i).Name, v.Field(i).Type(), v.Field(i).Interface())
		}
		r.Message = "ERROR:failed to convert data to base64"
	}

	r.Base64String = base64.StdEncoding.EncodeToString(byteArray)
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

type _ID struct {
	ID string `json:"id`
}
type _IDinfo struct {
	ArmorID string `json:"armorid"`
	AlphaID string `json:"alphaid"`
}

type _RecvRegisterData struct {
	Sign string  `json:"sign"`
	Data _IDinfo `json:"data"`
}

// Registered 用來給 client 註冊奧飛通行證資料
func Registered(w http.ResponseWriter, r *http.Request) {

	returnData := ReturnData{}
	returnData.State = 0

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvRegisterData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		fmt.Println(err)
		returnData.State = -1
		returnData.Message = "Registered:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	signData, _ := json.Marshal(recvData.Data)
	if returnData.signature(recvData.Sign, string(signData)) == false {
		fmt.Println(err)
		returnData.State = -2
		returnData.Message = "Registered:sign fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	currentID, err := strconv.Atoi(recvData.Data.ArmorID)
	if err != nil {
		returnData.State = -3
		returnData.Message = "Registered:igs id fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	result := mgodb.AlphaData.Registered(currentID, recvData.Data.AlphaID)

	resultString := result.Error()

	// returnData.data2Base64(resultString)
	returnData.Message = resultString
	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

type _RecvData struct {
	Sign string `json:"sign"`
	Data _ID    `json:"data"`
}

// PlayerinfoArmor 用來給 client 直接讀玩家資料使用
func PlayerinfoArmor(w http.ResponseWriter, r *http.Request) {

	returnData := ReturnData{}
	returnData.State = 0

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		fmt.Println(err)
		returnData.State = -1
		returnData.Message = "Playerinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	signData, _ := json.Marshal(recvData.Data)
	if returnData.signature(recvData.Sign, string(signData)) == false {
		fmt.Println(err)
		returnData.State = -2
		returnData.Message = "Playerinfo:sign fail"
		b, _ := json.Marshal(returnData)

		w.Write([]byte(b))
		return
	}

	currentID, err := strconv.Atoi(recvData.Data.ID)
	if err != nil {
		returnData.State = -3
		returnData.Message = "Playerinfo:igs id fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	info, err := mgodb.AlphaData.PlayerinfoArmor(currentID)

	if err != nil {
		returnData.State = -4
		returnData.Message = "Playerinfo:not found"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	returnData.data2Base64(info)
	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// PlayerinfoAlpha 用來給 client 直接讀玩家資料使用
func PlayerinfoAlpha(w http.ResponseWriter, r *http.Request) {

	returnData := ReturnData{}
	returnData.State = 0

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		fmt.Println(err)
		returnData.State = -1
		returnData.Message = "Playerinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	signData, _ := json.Marshal(recvData.Data)
	if returnData.signature(recvData.Sign, string(signData)) == false {
		fmt.Println(err)
		returnData.State = -2
		returnData.Message = "Playerinfo:sign fail"
		b, _ := json.Marshal(returnData)

		w.Write([]byte(b))
		return
	}

	info, err := mgodb.AlphaData.PlayerinfoAlpha(recvData.Data.ID)

	if err != nil {
		returnData.State = -4
		returnData.Message = "Playerinfo:not found"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	returnData.data2Base64(info)
	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// Rankinfo 用來給 client 直接讀取排名資料使用
func Rankinfo(w http.ResponseWriter, r *http.Request) {

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		fmt.Println(err)
		b, _ := json.Marshal("Rankinfo:json unmarshal fail")
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	returnData := ReturnData{}
	returnData.State = 0
	signData, _ := json.Marshal(recvData.Data)
	if returnData.signature(recvData.Sign, string(signData)) == false {
		fmt.Println(err)
		b, _ := json.Marshal("Rankinfo:sign fail")
		w.Write([]byte(b))
		return
	}

	index, err := strconv.Atoi(recvData.Data.ID)
	if err != nil {
		returnData.State = -3
		returnData.Message = "Rankinfo:activity id fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	rank, err := mgodb.AlphaData.Rankinfo(index)

	if err != nil {
		b, _ := json.Marshal("Rankinfo:not found")
		w.Write([]byte(b))
		return
	}

	returnData.data2Base64(rank)
	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))
}
