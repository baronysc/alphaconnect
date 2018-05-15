package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

type recvFrame struct {
	Sign string `json:"Sign,omitempty"`
	Data string `json:"Data,omitempty"`
}

type respFrame struct {
	Result  int    `json:"Result"`
	Sign    string `json:"Sign,omitempty"`
	Data    string `json:"Data,omitempty"`
	Message string `json:"Message,omitempty"`
}

//ParserFrame 解析框,回傳最後的參數是 errcode
func ParserFrame(data []byte, key string) ([]byte, error, int) {
	var err error
	recv := recvFrame{}
	err = json.Unmarshal(data, &recv)
	if err != nil {
		return nil, err, -1
	}

	if !CompareSign(key, recv.Sign, recv.Data) {
		err = errors.New("key Comparison error")
		return nil, err, -2
	}

	jdata, err := base64.StdEncoding.DecodeString(recv.Data)
	if err != nil {
		return nil, err, -3
	}
	return jdata, err, 0
}

//CombineFrame 組合框
func CombineFrame(jsondata []byte, key string, res int, msg string) []byte {
	resp := respFrame{}
	if res == 0 {
		resp.Data = base64.StdEncoding.EncodeToString(jsondata)
		resp.Sign = GenSign(key, resp.Data)
	}
	resp.Result = res
	resp.Message = msg
	jdata, _ := json.Marshal(resp)
	return jdata
}

//FailFrame 回傳異常
func FailFrame(res int, err error) []byte {
	resp := respFrame{}
	resp.Result = res
	resp.Message = fmt.Sprintln(err)
	jdata, _ := json.Marshal(resp)
	return jdata
}
