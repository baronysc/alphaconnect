package spec

import (
	"encoding/json"
	"log"
)

//
var expTable = `[
	[1,0],
	[2,100],
	[3,200],
	[4,300],
	[5,400],
	[6,500],
	[7,656],
	[8,868],
	[9,1136],
	[10,1460],
	[11,1840],
	[12,2276],
	[13,2768],
	[14,3316],
	[15,3920],
	[16,4580],
	[17,5296],
	[18,6068],
	[19,6896],
	[20,7780],
	[21,8720],
	[22,9716],
	[23,10768],
	[24,11876],
	[25,13040],
	[26,14260],
	[27,15536],
	[28,16868],
	[29,18256],
	[30,19700],
	[31,21200],
	[32,22700],
	[33,24200],
	[34,25700],
	[35,27200],
	[36,28700],
	[37,30200],
	[38,31700],
	[39,33200],
	[40,34700],
	[41,36200],
	[42,37700],
	[43,39200],
	[44,40700],
	[45,42200],
	[46,43700],
	[47,45200],
	[48,46700],
	[49,48200],
	[50,49700]
	]`

type ExpData struct {
	FA    float32 `json:"FA"`
	SKL   float32 `json:"SKL"`
	SPD   float32 `json:"SPD"`
	ATK   float32 `json:"ATK"`
	DEF   float32 `json:"DEF"`
	Total int32   `json:"Total"`
	//Type6 float32 `json:"Type6,omitempty"`
	//Type7 float32 `json:"Type7,omitempty"`
}

//expData 存放經驗值的陣列
var expData [][]interface{} = nil

//GetLevel 經驗值換算等級,回傳 小數2位數
func GetLevel(exp int) (float32, error) {
	var err error
	//第一次先把表讀到陣列
	if expData == nil {
		err := json.Unmarshal([]byte(expTable), &expData)
		if err != nil {
			log.Println(err)
			return 0, err
		}
	}

	//先設定最大等級
	level := float32(len(expData))
	for i := len(expData) - 1; i >= 0; i-- {
		score := int(expData[i][1].(float64))
		if exp >= score {
			//先取等大等級
			lv := float32(i + 1)
			//小於最大等級
			if i < len(expData)-1 {
				nowpct := expData[i][1].(float64)
				nextpct := expData[i+1][1].(float64)
				//會做多於的 int,float 主要是要去小數
				a := float32(int(((float64(exp)-nowpct)/(nextpct-nowpct))*100)) / 100
				lv += a
			}
			level = lv
			break
		}
	}
	return level, err
}

//GetAllLevel 取得所有等級
func GetAllLevel(e1 int32, e2 int32, e3 int32, e4 int32, e5 int32, e6 int32, e7 int32) ExpData {
	e := ExpData{}
	e.FA, _ = GetLevel(int(e1))
	e.SKL, _ = GetLevel(int(e2))
	e.SPD, _ = GetLevel(int(e3))
	e.ATK, _ = GetLevel(int(e4))
	e.DEF, _ = GetLevel(int(e5))
	e.Total = int32(e.FA) + int32(e.SKL) + int32(e.SPD) + int32(e.ATK) + int32(e.DEF) - 4
	//e.Type6, _ = GetLevel(int(l5))
	//e.Type7, _ = GetLevel(int(l5))
	return e
}
