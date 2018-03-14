package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"env"
	"errors"
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
	Sign string `json:"Sign"`
	Data string `json:"Data"`
}

// AlphaKey 用來提供給 hmac 轉換，並驗證傳值的正確性
const AlphaKey string = "OQrdcqpv26hBr8ef"

func (r *ReturnData) setSignature(_Data string) {

	key := []byte(AlphaKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(_Data))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	r.Sign = sign
}

// compareSignature 簽章判斷
func (r *ReturnData) compareSignature(_Sign string, _Data string) bool {

	key := []byte(AlphaKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(_Data))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if strings.Compare(sign, _Sign) != 0 {
		return false
	}

	return true
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
	}

	r.Data = base64.StdEncoding.EncodeToString(byteArray)
}

func (r *ReturnData) decodeArmorID(_ArmorID string) (int, error) {

	var err error

	if len(_ArmorID) != 12 { // 編碼不足12碼就表示資料錯誤
		err = errors.New("id size not enough")
		return 1, err
	}

	// 編碼轉換
	checksum, err := strconv.Atoi(_ArmorID[len(_ArmorID)-2 : len(_ArmorID)])
	if err != nil {
		err = errors.New("check sum get fail")
		return 2, err
	}
	sOriginID := _ArmorID[:len(_ArmorID)-2]
	iOriginID, err := strconv.Atoi(sOriginID)
	if err != nil {
		err = errors.New("id string to int fail")
		return 3, err
	}

	temp := 0
	for {
		temp += iOriginID & 0x0f
		iOriginID = iOriginID >> 8

		if iOriginID <= 0 {
			break
		}
	}
	if temp != checksum {
		err = errors.New("check sum compare fail")
		return 4, err
	}

	currentID, _ := strconv.Atoi(sOriginID)
	if err != nil {
		err = errors.New("current id string to ing fail")
		return 5, err
	}

	return currentID, nil
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
	ID string `json:"id"`
}
type _IDinfo struct {
	ArmorID string `json:"armorid"`
	AlphaID string `json:"alphaid"`
}

type _RecvData struct {
	Sign string `json:"sign"`
	Data string `json:"data"`
}

// Registered 用來給 client 註冊奧飛通行證資料
func Registered(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha registered ")

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.Data = "Registered:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		msg.Log("Registered:sign fail")
		returnData.Data = "Registered:sign fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	id := _IDinfo{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.Data = "Registered:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	currentID, err := returnData.decodeArmorID(id.ArmorID)
	if err != nil {
		returnData.Data = "Registered:igs id fail:" + err.Error()
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	msg.Log("register current id:", currentID, ", alpha id:", id.AlphaID)
	result := mgodb.AlphaData.Registered(currentID, id.AlphaID)

	returnData.data2Base64(result.Error())
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// PlayerinfoArmor 用來給 client 直接讀玩家資料使用
func PlayerinfoArmor(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search playerinfot from Armor account")

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.Data = "Playerinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		msg.Log("Registered:sign fail")
		returnData.Data = "Playerinfo:sign fail"
		b, _ := json.Marshal(returnData)

		w.Write([]byte(b))
		return
	}

	id := _ID{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.Data = "Playerinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	VerifyError := ""

	if len(id.ID) != 12 { // 編碼不足12碼就表示資料錯誤
		VerifyError = "1"
		msg.Log("Verify:", VerifyError)
		returnData.Data = "Playerinfo:igs id fail:" + VerifyError
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 編碼轉換
	currentID, err := returnData.decodeArmorID(id.ID)
	if err != nil {
		returnData.Data = "Playerinfo:igs id fail:" + err.Error()
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	msg.Log("search player id:", currentID)
	info, err := mgodb.AlphaData.PlayerinfoArmor(currentID)

	if err != nil {
		returnData.Data = "Playerinfo:not found"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	returnData.data2Base64(info)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// PlayerinfoAlpha 用來給 client 直接讀玩家資料使用
func PlayerinfoAlpha(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search playerinfot from Alpha account")

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.Data = "Playerinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		msg.Log("Registered:sign fail")
		returnData.Data = "Playerinfo:sign fail"
		b, _ := json.Marshal(returnData)

		w.Write([]byte(b))
		return
	}

	id := _ID{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.Data = "Playerinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	msg.Log("search alpha id:", id.ID)

	info, err := mgodb.AlphaData.PlayerinfoAlpha(id.ID)

	if err != nil {
		msg.Log(err)
		returnData.Data = "Playerinfo:not found"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	returnData.data2Base64(info)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// Rankinfo 用來給 client 直接讀取排名資料使用
func Rankinfo(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search ranking data")

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.Data = "Rankinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		msg.Log("Registered:sign fail")
		returnData.Data = "Rankinfo:sign fail"
		b, _ := json.Marshal(returnData)

		w.Write([]byte(b))
		return
	}

	id := _ID{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.Data = "Rankinfo:json unmarshal fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	index, err := strconv.Atoi(id.ID)
	if err != nil {
		returnData.Data = "Rankinfo:activity id fail"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))
		return
	}

	msg.Log("search ranking index:", index)
	rank, err := mgodb.AlphaData.Rankinfo(index)

	if err != nil {
		returnData.Data = "Rankinfo:not found"
		b, _ := json.Marshal(returnData)
		w.Write([]byte(b))

		return
	}

	returnData.data2Base64(rank)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))
}
