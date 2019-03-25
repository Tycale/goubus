package goubus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

//UbusResponseCode represent the status code from JSON-RPC Call
type UbusResponseCode float64

//Represents enum ubus_msg_status from https://git.openwrt.org/?p=project/ubus.git;a=blob;f=ubusmsg.h;h=398b126b6dc01833937749a110181ea0debb1476;hb=HEAD
const (
	UbusStatusOK               UbusResponseCode = 0
	UbusStatusInvalidCommand   UbusResponseCode = 1
	UbusStatusInvalidArgument  UbusResponseCode = 2
	UbusStatusMethodNotFound   UbusResponseCode = 3
	UbusStatusNotFound         UbusResponseCode = 4
	UbusStatusNoData           UbusResponseCode = 5
	UbusStatusPermissionDenied UbusResponseCode = 6
	UbusStatusTimeout          UbusResponseCode = 7
	UbusStatusNotSupported     UbusResponseCode = 8
	UbusStatusUnknownError     UbusResponseCode = 9
	UbusStatusConnectionFailed UbusResponseCode = 10
	UbusStatusLast             UbusResponseCode = 11
)

//Ubus represents information to JSON-RPC Interaction with router
type Ubus struct {
	Username string
	Password string
	URL      string
	AuthData UbusAuthData
}

//UbusResponse represents a response from JSON-RPC
type UbusResponse struct {
	JSONRPC          string
	ID               int
	Result           interface{}
	UbusResponseCode UbusResponseCode
}

//LoginCheck check if login RPC Session id has expired
func (u *Ubus) LoginCheck() error {
	var err error
	var i uint8
	for start := time.Now(); time.Since(start) < 3*time.Second; {
		if u.AuthData.ExpireTime.Before(time.Now()) {
			_, err = u.AuthLogin()
			if err == nil {
				break
			}
		} else {
			break
		}
		i++
		time.Sleep(time.Second)
	}
	if i == 2 {
		return errors.New("Login Timeout")
	}
	return nil
}

//Call do a call to Json-RPC to get/set information
func (u *Ubus) Call(jsonStr []byte) (UbusResponse, error) {
	req, err := http.NewRequest("POST", u.URL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		return UbusResponse{}, errors.New(resp.Status)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	result := UbusResponse{}
	json.Unmarshal([]byte(body), &result)
	//Workaround cause response code not contempled by unmarshal function
	result.UbusResponseCode = UbusResponseCode(result.Result.([]interface{})[0].(float64))
	//Workaround to get UbusData cause the structure of this array has a problem with unmarshal
	if result.UbusResponseCode == UbusStatusOK {
		return result, nil
	}
	return UbusResponse{}, fmt.Errorf("Ubus Status Failed: %v", result.UbusResponseCode)
}
