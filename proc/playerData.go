package proc

import (
	"alphaconnect/noti"
	"alphaconnect/spec"
	"alphaconnect/util"
	"encoding/json"
	"io/ioutil"
	"log"
	"mgodb"
	"net/http"
	"strconv"
)

type recvID struct {
	ID string `json:"ID"`
}

//PlayerData 回傳的結構
type PlayerData struct {
	Name            string               `json:"name"`
	CreateTime      int64                `json:"createTime"`
	RegisteredTime  int64                `json:"registeredTime"`
	AlphaID         string               `json:"alphaID"`
	LastTime        int64                `json:"lastTime"`
	InheritCnt      int32                `json:"inheritCnt"`
	ExpData         spec.ExpData         `json:"expData"`
	Title           string               `json:"title"`
	TitleOwn        []string             `json:"titleOwn"`
	AchievementData spec.AchievementData `json:"achievementData"`
}

// HandlerPlayerDataArmor 處理個人資料
func HandlerPlayerDataArmor(w http.ResponseWriter, r *http.Request) {
	originSource, _ := ioutil.ReadAll(r.Body)
	log.Println("接收到 PlayerDataArmor 請求:" + string(originSource))
	//log.Printf("origin:%v\n", string(originSource))

	//先做變數的宣告(如果使用GOTO 不能在 GOTO以下使用臨時變數)
	var recvJSON []byte          //接收後去框資料
	var respData []byte          //回傳資料
	var currentID int            //玩家目前ID
	var info *mgodb.G_PlayerInfo //從 DB取出的資料
	var playerData []byte        //要回傳的玩家資料
	var rID recvID               //取出玩家ID

	recvJSON, err, errcode := util.ParserFrame(originSource, noti.AlphaKey)
	if err != nil {
		respData = util.FailFrame(errcode, err)
		goto L
	}

	err = json.Unmarshal(recvJSON, &rID)
	if err != nil {
		respData = util.FailFrame(-1, err)
		goto L
	}

	currentID, err = util.CheckArmorID(rID.ID)
	if err != nil {
		respData = util.FailFrame(-4, err)
		goto L
	}

	info, err = mgodb.AlphaData.PlayerinfoArmor(currentID)
	if err != nil {
		respData = util.FailFrame(-5, err)
		goto L
	}

	playerData = GetPlayerData(info)
	respData = util.CombineFrame(playerData, noti.AlphaKey, 0, "")

L:
	w.Write(respData)
}

// HandlerPlayerDataArmor 處理個人資料
func HandlerPlayerDataCreateID(w http.ResponseWriter, r *http.Request) {
	originSource, _ := ioutil.ReadAll(r.Body)
	log.Println("接收到 PlayerDataArmor 請求:" + string(originSource))
	//log.Printf("origin:%v\n", string(originSource))

	//先做變數的宣告(如果使用GOTO 不能在 GOTO以下使用臨時變數)
	var recvJSON []byte          //接收後去框資料
	var respData []byte          //回傳資料
	var createID int             //玩家目前ID
	var info *mgodb.G_PlayerInfo //從 DB取出的資料
	var playerData []byte        //要回傳的玩家資料
	var rID recvID               //取出玩家ID

	recvJSON, err, errcode := util.ParserFrame(originSource, noti.AlphaKey)
	if err != nil {
		respData = util.FailFrame(errcode, err)
		goto L
	}

	err = json.Unmarshal(recvJSON, &rID)
	if err != nil {
		respData = util.FailFrame(-1, err)
		goto L
	}

	log.Println(rID.ID)
	createID, err = strconv.Atoi(rID.ID)
	if err != nil {
		respData = util.FailFrame(-4, err)
		goto L
	}

	/*
		createID, err = util.CheckrAmorID(rID.ID)
		if err != nil {
			respData = util.FailFrame(-1, err)
			goto L
		}
	*/

	info, err = mgodb.AlphaData.PlayerinfoCreateID(createID)
	if err != nil {
		respData = util.FailFrame(-5, err)
		goto L
	}

	playerData = GetPlayerData(info)
	respData = util.CombineFrame(playerData, noti.AlphaKey, 0, "")

L:
	w.Write(respData)
}

//HandlerPlayerDataAlpha 使用奧飛通行證取得資料,必須先綁定過後才能用
func HandlerPlayerDataAlpha(w http.ResponseWriter, r *http.Request) {
	originSource, _ := ioutil.ReadAll(r.Body)
	log.Println("接收到 PlayerDataAlpha 請求:" + string(originSource))

	//先做變數的宣告(如果使用GOTO 不能在 GOTO以下使用臨時變數)
	var recvJSON []byte          //接收後去框資料
	var respData []byte          //回傳資料
	var alphaID string           //奧飛通行証
	var info *mgodb.G_PlayerInfo //從 DB取出的資料
	var playerData []byte        //要回傳的玩家資料
	var rID recvID               //取出玩家ID

	recvJSON, err, errcode := util.ParserFrame(originSource, noti.AlphaKey)
	if err != nil {
		respData = util.FailFrame(errcode, err)
		goto L
	}

	err = json.Unmarshal(recvJSON, &rID)
	if err != nil {
		respData = util.FailFrame(-1, err)
		goto L
	}

	err = json.Unmarshal(recvJSON, &rID)
	if err != nil {
		respData = util.FailFrame(-4, err)
		goto L
	}

	//奧飛通行証
	alphaID = rID.ID
	info, err = mgodb.AlphaData.PlayerinfoAlpha(alphaID)
	if err != nil {
		respData = util.FailFrame(-5, err)
		goto L
	}

	playerData = GetPlayerData(info)
	respData = util.CombineFrame(playerData, noti.AlphaKey, 0, "")

L:
	w.Write(respData)
}

//GetPlayerData 取得玩家的數據
func GetPlayerData(info *mgodb.G_PlayerInfo) []byte {
	//填入資料
	pd := PlayerData{}
	pd.Name = info.RFIDCardData.AccountData.Name
	pd.CreateTime = info.RFIDCardData.AccountData.CreateTime
	pd.LastTime = info.LastOnLineTime
	pd.InheritCnt = info.RFIDCardData.AccountData.InheritCount
	pd.RegisteredTime = info.RegisteredTime
	pd.AlphaID = info.AlphaID

	//寫入經驗值
	e1 := info.RFIDCardData.GetExpData().GetType1()
	e2 := info.RFIDCardData.GetExpData().GetType2()
	e3 := info.RFIDCardData.GetExpData().GetType3()
	e4 := info.RFIDCardData.GetExpData().GetType4()
	e5 := info.RFIDCardData.GetExpData().GetType5()
	e6 := info.RFIDCardData.GetExpData().GetType6()
	e7 := info.RFIDCardData.GetExpData().GetType7()
	pd.ExpData = spec.GetAllLevel(e1, e2, e3, e4, e5, e6, e7)

	//寫入稱號
	pd.Title, _ = spec.GetTitle(int(info.RFIDCardData.TitleData.GetIndex()))
	t1 := uint(info.RFIDCardData.TitleData.GetStatus())
	t2 := uint(info.RFIDCardData.TitleData.GetStatusEx1())
	t3 := uint(info.RFIDCardData.TitleData.GetStatusEx2())
	t4 := uint(info.RFIDCardData.TitleData.GetStatusEx3())
	jd, _ := spec.GetOwnTitle(t1, t2, t3, t4)
	_ = json.Unmarshal([]byte(jd), &pd.TitleOwn)

	//寫入成就
	pd.AchievementData = spec.GetAchievement(info.RFIDCardData.GetAchievementData().GetData())

	jdata, _ := json.Marshal(pd)
	return jdata

}
