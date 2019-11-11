package reporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type SmsResponseData struct {
	Sid   string `json:"sid"`
	Fee   int    `json:"fee"`
	Count int    `json:"count"`
}

type SmsResponse struct {
	Reason    string           `json:"reason"`
	Result    *SmsResponseData `json:"result"`
	ErrorCode int              `json:"error_code"`
}

type SMS struct {
	AppKey string
}

func (s SMS) SMSSend(mobile, tplId string, data map[string]interface{}) error {
	stringList := []string{}
	for key, val := range data {
		stringList = append(stringList, fmt.Sprintf("#%s#=%v", key, val))
	}
	tplValue := strings.Join(stringList, "&")
	param := url.Values{}
	param.Set("mobile", mobile)
	param.Set("tpl_id", tplId)
	param.Set("tpl_value", tplValue)
	param.Set("key", s.AppKey)
	apiURL := "http://v.juhe.cn/sms/send"
	var (
		respData []byte
		err      error
	)
	if respData, err = s.get(apiURL, param); err != nil {
		log.Printf("Juhe SMS send err: %v; param: %s", err, param.Encode())
		return nil
	}
	var resp SmsResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		log.Printf("Juhe SMS send err: %v; data: %s", err, string(respData))
		return err
	}
	if resp.ErrorCode != 0 {
		log.Printf("Juhe SMS send err: %v; data: %s", err, string(respData))
		return errors.New(resp.Reason)
	}
	log.Printf("Juhe SMS send info param: %v; data: %s", param.Encode(), string(respData))
	return nil
}

func (s SMS) get(apiURL string, params url.Values) ([]byte, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	u.RawQuery = params.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
