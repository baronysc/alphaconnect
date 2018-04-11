package spec

//arr[0]    --遊玩次數(A01)
//arr[1]    --未定
//arr[2]    --發動復活成功次數(A03)
//arr[3]    --完美提升鬥志次數(A04)
//arr[4]    --發動暴擊次數(A05)
//arr[5]    --發動幸運次數(A06)
//arr[6]    --發動必殺技次數(A07)

//AchievementData 逹成成就
type AchievementData struct {
	A01 int32 `json:"A01"`           //遊玩次數
	A02 int32 `json:"A02,omitempty"` //未定,所以不傳
	A03 int32 `json:"A03"`           //發動復活成功次數
	A04 int32 `json:"A04"`           //完美提升鬥志次數
	A05 int32 `json:"A05"`           //發動暴擊次數
	A06 int32 `json:"A06"`           //發動幸運次數
	A07 int32 `json:"A07"`           //發動必殺技次數
}

/*
func GetAchievement(arr []int) (string, error) {
	var err error
	m := make(map[string]int)
	m["A01"] = arr[0] //遊玩次數
	//m["A02"] = arr[1] //未定
	m["A03"] = arr[2] //發動復活成功次數
	m["A04"] = arr[3] //完美提升鬥志次數
	m["A05"] = arr[4] //發動暴擊次數
	m["A06"] = arr[5] //發動幸運次數
	m["A07"] = arr[6] //發動必殺技次數

	jdata, err := json.Marshal(m)
	return string(jdata), err
}
*/

// GetAchievement 回傳成就
func GetAchievement(arr []int32) AchievementData {
	a := AchievementData{}
	a.A01 = arr[0]
	a.A02 = arr[1]
	a.A03 = arr[2]
	a.A04 = arr[3]
	a.A05 = arr[4]
	a.A06 = arr[5]
	a.A07 = arr[6]
	return a
}
