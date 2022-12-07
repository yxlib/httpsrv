// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/yxlib/yx"
)

var (
	ErrNotSupportMethod = errors.New("not support http method")
)

var ec = yx.NewErrCatcher("httpsrv.util")
var logger = yx.NewLogger("httpsrv.util")

func DefaultRead(req *http.Request, cfg *Config) (*Request, error) {
	var err error = nil
	defer ec.DeferThrow("DefaultRead", &err)

	logger.Detail(yx.LOG_LV_INFO, []string{"\n"})
	logger.I("## Http Request Start ##")

	// raw data
	reqData, err := GetReqData(req)
	if err != nil {
		return nil, err
	}

	logger.Detail(yx.LOG_LV_DEBUG, []string{"[R] Request Raw: ", reqData, "\n"})

	// parse query
	val, err := ParseQuery(reqData)
	if err != nil {
		return nil, err
	}

	reqObj := &Request{}

	reqObj.Token = val.Get(cfg.TokenField)
	// logger.D("Unescape Token: ", reqObj.Token)

	reqObj.Pattern = req.URL.Path
	// logger.I("Pattern: ", reqObj.Pattern)

	reqObj.Opr = val.Get(cfg.OprField)
	// logger.I("Operation: ", reqObj.Opr)

	snoStr := val.Get(cfg.SerialNoField)
	sno, err := strconv.ParseUint(snoStr, 10, 16)
	if err == nil {
		reqObj.SerialNo = uint16(sno)
		// logger.I("SerialNo: ", reqObj.SerialNo)
	}

	reqObj.Params = val.Get(cfg.ParamsField)
	// logger.D("Unescape Data: ", reqObj.Params)

	log := fmt.Sprint("[0] Pattern = ", reqObj.Pattern, ", Opr = ", reqObj.Opr, ", SNo = ", reqObj.SerialNo, "\n")
	logger.Detail(yx.LOG_LV_INFO, []string{log})
	logger.Detail(yx.LOG_LV_DEBUG, []string{"[1] Token: ", reqObj.Token, "\n", "[2] Data: ", reqObj.Params, "\n"})

	return reqObj, nil
}

func DefaultWrite(writer http.ResponseWriter, cfg *Config, respObj *Response, err error) error {
	writer.Header().Set("content-type", "application/x-www-form-urlencoded")

	respResult := respObj.Result
	respCode := respObj.Code
	if err != nil {
		respResult = "\"" + err.Error() + "\""
	}

	respResult = EncodeURIComponent(respResult)
	// respData := cfg.CodeField + "=" + strconv.Itoa(int(respCode)) + "&" + cfg.ResultField + "=" + string(respResult)
	respData := fmt.Sprintf("%s=%s&%s=%d&%s=%d&%s=%s", cfg.OprField, respObj.Opr, cfg.SerialNoField, respObj.SerialNo, cfg.CodeField, respCode, cfg.ResultField, respResult)

	logger.Detail(yx.LOG_LV_DEBUG, []string{"[0] Response Raw: ", respData, "\n"})
	logger.Detail(yx.LOG_LV_INFO, []string{"\n"})

	_, errWrite := writer.Write([]byte(respData))
	return ec.Throw("DefaultWrite", errWrite)
}

func GetReqData(req *http.Request) (string, error) {
	reqData := ""
	if req.Method == http.MethodGet {
		reqData = req.URL.RawQuery

	} else if req.Method == http.MethodPost {
		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return "", err
		}

		reqData = string(d)

	} else {
		return "", ErrNotSupportMethod
	}

	return reqData, nil
}

func ParseQuery(query string) (url.Values, error) {
	var err error = nil
	m := make(url.Values)

	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := DecodeURIComponent(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = DecodeURIComponent(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = append(m[key], value)
	}
	return m, err
}

const upperhex = "0123456789ABCDEF"

func EncodeURIComponent(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}

	if hexCount == 0 {
		return s
	}

	var buf [64]byte
	var t []byte

	required := len(s) + 2*hexCount
	if required <= len(buf) {
		t = buf[:required]
	} else {
		t = make([]byte, required)
	}

	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case shouldEscape(c):
			t[j] = '%'
			t[j+1] = upperhex[c>>4]
			t[j+2] = upperhex[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func DecodeURIComponent(s string) (string, error) {
	return url.PathUnescape(s)
}

func shouldEscape(c byte) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
		return false
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '!', '*', '\'', '(', ')':
		return false
	}

	// Everything else must be escaped.
	return true
}
