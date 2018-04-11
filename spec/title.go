package spec

import (
	"encoding/json"
	"log"
)

var titleTable = `[
	[1,"战士",1],
	[2,"斗士",2],
	[3,"勇士",3],
	[4,"英雄",3],
	[5,"真英雄",3],
	[6,"盖世英雄",4],
	[7,"天下无敌",4],
	[8,"霸王",4],
	[9,"魔王",4],
	[10,"斗神",4],
	[11,"新人",1],
	[12,"经验丰富",1],
	[13,"专家",2],
	[14,"历战勇士",3],
	[15,"战斗大师",4],
	[16,"讯息守卫",2],
	[17,"守护神",3],
	[18,"追击者",2],
	[19,"激斗者",2],
	[20,"锐变",2],
	[21,"超进化",3],
	[22,"极限进化",4],
	[23,"变幻自在",3],
	[24,"不灭斗志",3],
	[25,"眼明手快",3],
	[26,"暴发力",3],
	[27,"幸运",2],
	[28,"超幸运",3],
	[29,"终结者",2],
	[30,"正法不义",3],
	[31,"进化者",2],
	[32,"神威降临",2]
	]`

var titleData [][]interface{} = nil

//loadData 載入資料到陣列
func loadData() error {
	var err error
	if titleData == nil {
		err := json.Unmarshal([]byte(titleTable), &titleData)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return err
}

//GetTitle 取得稱號
func GetTitle(index int) (string, error) {
	var err error

	err = loadData()
	if err != nil {
		return "無", err
	}

	for i := 0; i < len(titleData); i++ {
		titleIdx := int(titleData[i][0].(float64))
		if index == titleIdx {
			return titleData[i][1].(string), nil
		}
	}
	return "", err
}

//GetOwnTitle 取得擁有的稱號,回傳 Json["战士","斗士","真英雄","盖世英雄","天下无敌"]
func GetOwnTitle(t1 uint, t2 uint, t3 uint, t4 uint) (string, error) {
	var err error
	var arr []string
	//1~30號
	b := uint(0x01)
	for i := 1; i <= 120; i++ {
		if i < 31 {
			s := uint(i - 1)
			if (t1 & (b << s)) > 0 {
				t, err := GetTitle(i)
				if err != nil {
					return "", err
				}
				arr = append(arr, t)
			}
		} else if i < 61 {
			s := uint(i - 31)
			if (t2 & (b << s)) > 0 {
				t, err := GetTitle(i)
				if err != nil {
					return "", err
				}
				arr = append(arr, t)
			}
		} else if i < 91 {
			s := uint(i - 61)
			if (t3 & (b << s)) > 0 {
				t, err := GetTitle(i)
				if err != nil {
					return "", err
				}
				arr = append(arr, t)
			}
		} else {
			s := uint(i - 91)
			if (t4 & (b << s)) > 0 {
				t, err := GetTitle(i)
				if err != nil {
					return "", err
				}
				arr = append(arr, t)
			}
		}
	}

	jdata, err := json.Marshal(arr)
	return string(jdata), err
}
