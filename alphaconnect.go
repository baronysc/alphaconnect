package main

//go get github.com/fatih/color
//go get -u github.com/davecgh/go-spew/spew

import (
	"alphaconnect/proc"
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
		http.HandleFunc("/RegisteredStatus", RegisteredStatus)
		http.HandleFunc("/Playerinfo_Alpha", PlayerinfoAlpha)
		http.HandleFunc("/Playerinfo_Armor", PlayerinfoArmor)
		http.HandleFunc("/ActivityStatus", ActivityStatus)
		http.HandleFunc("/Rankinfo", Rankinfo)
		http.HandleFunc("/Rankinfo_Armor", RankinfoArmor)
		http.HandleFunc("/Rankinfo_Alpha", RankinfoAlpha)

		http.HandleFunc("/StoreList", StoreList)

		http.HandleFunc("/PlayerGameList", PlayerGameList)

		//取得個人資料(以下2個是要取代 Playerinfo_Alpha,Playerinfo_Armor,但舊的先不能刪除,主要目前奧飛有使用到)
		http.HandleFunc("/PlayerData_Armor", proc.HandlerPlayerDataArmor)
		http.HandleFunc("/PlayerData_Alpha", proc.HandlerPlayerDataAlpha)

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
	Result  int    `json:"Result"`
	Sign    string `json:"Sign,omitempty"`
	Data    string `json:"Data,omitempty"`
	Message string `json:"Message,omitempty"`
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

func (r *ReturnData) errorMessage(w http.ResponseWriter, _ErrorNo int, _ErrorMessage string) {
	r.Result = _ErrorNo
	r.Message = _ErrorMessage
	msg.Log(r.Message)
	b, err := json.Marshal(r)
	if err != nil {
		w.Write([]byte("Message:base error message json marshal fail"))
	} else {
		w.Write([]byte(b))
	}
}

func (r *ReturnData) sendCorrectData(w http.ResponseWriter) {

	b, err := json.Marshal(r)
	if err != nil {
		w.Write([]byte("Message:send correct data json marshal fail"))
	} else {
		w.Write([]byte(b))
	}
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

func getIP(r *http.Request) string {

	clientIP := r.Header.Get("x-forwarded-for")

	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	return clientIP
}

// JokeBack 用來給 client 直接讀取判斷 ip 使用
func JokeBack(w http.ResponseWriter, r *http.Request) {

	content := _ServerSayHello{
		ClientIP: getIP(r),
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
type _IDRankinfo struct {
	ID    string `json:"id"`
	Index string `json:"index"`
}

type _RecvData struct {
	Sign string `json:"sign"`
	Data string `json:"data"`
}

// Registered 用來給 client 註冊奧飛通行證資料
func Registered(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha registered ")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.errorMessage(w, -1, "Registered:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "Registered:sign fail")
		return
	}

	id := _IDinfo{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "Registered:json unmarshal fail")
		return
	}

	currentID, err := returnData.decodeArmorID(id.ArmorID)
	if err != nil {
		returnData.errorMessage(w, -4, "Registered:igs id fail, "+err.Error())
		return
	}

	msg.Log("register current id:", currentID, ", alpha id:", id.AlphaID)
	result := mgodb.AlphaData.Registered(currentID, id.AlphaID)

	if result.Error() == "build" {
		returnData.Result = 0
	} else if result.Error() == "replace" {
		returnData.Result = 1
	} else {
		returnData.errorMessage(w, -5, "Registered:not found")
		return
	}

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// RegisteredStatus 用來給 client 註冊奧飛通行證資料
func RegisteredStatus(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha registered status")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.errorMessage(w, -1, "RegisteredStatus:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "RegisteredStatus:sign fail")
		return
	}

	id := _ID{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "RegisteredStatus:json unmarshal fail")
		return
	}

	msg.Log("register current alpha id:", id.ID)

	_, err = mgodb.AlphaData.PlayerinfoAlpha(id.ID)
	if err != nil {
		returnData.errorMessage(w, -4, "RegisteredStatus:not found")
		return
	}

	returnData.Result = 0

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// PlayerinfoArmor 用來給 client 直接讀玩家資料使用
func PlayerinfoArmor(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search playerinfo from Armor account")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		returnData.errorMessage(w, -1, "Playerinfo:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "Playerinfo:sign fail")
		return
	}

	id := _ID{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "Playerinfo:json unmarshal fail")
		return
	}

	// 編碼轉換
	currentID, err := returnData.decodeArmorID(id.ID)
	if err != nil {
		returnData.errorMessage(w, -4, "Playerinfo:igs id fail:"+err.Error())
		return
	}

	msg.Log("search armor player id:", currentID)
	info, err := mgodb.AlphaData.PlayerinfoArmor(currentID)

	if err != nil {
		returnData.errorMessage(w, -5, "Playerinfo:not found")
		return
	}

	returnData.Result = 0
	returnData.data2Base64(info)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// PlayerinfoAlpha 用來給 client 直接讀玩家資料使用
func PlayerinfoAlpha(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search playerinfo from Alpha account")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)

		returnData.errorMessage(w, -1, "Playerinfo:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "Playerinfo:sign fail")
		return
	}

	id := _ID{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "Playerinfo:json unmarshal fail")
		return
	}

	msg.Log("search alpha player id:", id.ID)

	info, err := mgodb.AlphaData.PlayerinfoAlpha(id.ID)
	if err != nil {
		returnData.errorMessage(w, -4, "Playerinfo:not found")
		return
	}

	returnData.Result = 0
	returnData.data2Base64(info)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// ActivityInfo 活動資料
type ActivityInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Start int64  `json:"start"`
	End   int64  `json:"end"`
}

// ActivityStatus 用來給 client 直接讀取排名資料使用
func ActivityStatus(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search activity information")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}
	ContentInfo := mgodb.AlphaData.ActivityStatus()

	if ContentInfo.Open == false {
		returnData.errorMessage(w, -1, "ActivityStatus:no acticity")
		return
	}

	info := ActivityInfo{
		ID:    ContentInfo.Item,
		Name:  ContentInfo.Title,
		Start: ContentInfo.Start,
		End:   ContentInfo.End,
	}

	returnData.Result = 0
	returnData.data2Base64(info)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// Rankinfo 用來給 client 直接讀取排名資料使用
func Rankinfo(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search ranking data")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.errorMessage(w, -1, "Rankinfo:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "Rankinfo:sign fail")
		return
	}

	id := _ID{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "Rankinfo:json unmarshal fail")
		return
	}

	index, err := strconv.Atoi(id.ID)
	if err != nil {
		returnData.errorMessage(w, -4, "Rankinfo:activity id fail")
		return
	}

	msg.Log("search ranking index:", index)
	rank, err := mgodb.AlphaData.Rankinfo(index)
	if err != nil {
		returnData.errorMessage(w, -5, "Rankinfo:not found")
		return
	}

	returnData.Result = 0
	returnData.data2Base64(rank)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))
}

// RankinfoArmor 用來給 client 依玩家資料取排名使用
func RankinfoArmor(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search player rank for armor id")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		returnData.errorMessage(w, -1, "Rankinfo:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "Rankinfo:sign fail")
		return
	}

	id := _IDRankinfo{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "Rankinfo:json unmarshal fail")
		return
	}

	currentID, err := returnData.decodeArmorID(id.ID)
	if err != nil {
		returnData.errorMessage(w, -4, "Rankinfo:igs id decode fail, "+err.Error())
		return
	}

	msg.Log("search armor player id:", currentID)

	index, err := strconv.Atoi(id.Index)
	if err != nil {
		returnData.errorMessage(w, -5, "Rankinfo:activity index fail")
		return
	}

	msg.Log("search armor player id:", currentID)
	info, err := mgodb.AlphaData.PlayerinfoArmor(currentID)
	if err != nil {
		returnData.errorMessage(w, -6, "Rankinfo:not found")
		return
	}

	rank, err := mgodb.AlphaData.RankinfoFromPlayer(index, info.CreateId)
	if err != nil {
		returnData.errorMessage(w, -7, "Rankinfo:not found")
		return
	}

	returnData.Result = 0
	returnData.data2Base64(rank)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))
}

// RankinfoAlpha 用來給 client 依玩家資料取排名使用
func RankinfoAlpha(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search player rank for alpha id")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		msg.Log(err)
		returnData.errorMessage(w, -1, "Rankinfo:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "Rankinfo:sign fail")
		return
	}

	id := _IDRankinfo{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "Rankinfo:json unmarshal fail")
		return
	}

	msg.Log("search alpha player id:", id.ID)

	info, err := mgodb.AlphaData.PlayerinfoAlpha(id.ID)
	if err != nil {
		msg.Log(err)
		returnData.errorMessage(w, -4, "Rankinfo:not found")
		return
	}

	msg.Log("get armor player id:", info.CurrentId)

	index, err := strconv.Atoi(id.Index)
	if err != nil {
		returnData.errorMessage(w, -5, "Rankinfo:activity index fail")
		return
	}

	rank, err := mgodb.AlphaData.RankinfoFromPlayer(index, info.CurrentId)
	if err != nil {
		returnData.errorMessage(w, -6, "Rankinfo:not found")
		return
	}

	returnData.Result = 0
	returnData.data2Base64(rank)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// StoreList 提供給奧飛取得店家機台資料
func StoreList(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide alpha search store information")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	totalList := mgodb.AlphaData.StoreList()
	if totalList == nil {
		returnData.errorMessage(w, -1, "StoreList:not found")
		return
	}

	returnData.Result = 0
	returnData.data2Base64(totalList)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}

// PlayerGameList 用來給 client 依玩家資料取排名使用
func PlayerGameList(w http.ResponseWriter, r *http.Request) {

	msg.Log("Provide search player game list for armor id")
	msg.Log("From:", getIP(r))

	returnData := ReturnData{}

	originSource, _ := ioutil.ReadAll(r.Body)

	recvData := _RecvData{}
	err := json.Unmarshal(originSource, &recvData)
	if err != nil {
		returnData.errorMessage(w, -1, "PlayerGameList:json unmarshal fail")
		return
	}

	// 簽章認證
	if returnData.compareSignature(recvData.Sign, string(recvData.Data)) == false {
		returnData.errorMessage(w, -2, "PlayerGameList:sign fail")
		return
	}

	id := _IDRankinfo{}

	byteArray, _ := base64.StdEncoding.DecodeString(recvData.Data)
	err = json.Unmarshal(byteArray, &id)
	if err != nil {
		returnData.errorMessage(w, -3, "PlayerGameList:json unmarshal fail")
		return
	}

	// 必須使用 create id 否則會取不到資料
	createID, err := returnData.decodeArmorID(id.ID)
	if err != nil {
		returnData.errorMessage(w, -4, "PlayerGameList:igs id decode fail:"+err.Error())
		return
	}

	msg.Log("search armor player id:", createID)

	playerList, err := mgodb.AlphaData.PlayerGameList(createID)
	if err != nil {
		returnData.errorMessage(w, -5, "PlayerGameList:get player game list fail:"+err.Error())
		return
	}

	returnData.Result = 0
	returnData.data2Base64(playerList)
	returnData.setSignature(returnData.Data)

	b, _ := json.Marshal(returnData)

	w.Write([]byte(b))

}
