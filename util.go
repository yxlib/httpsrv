// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

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
