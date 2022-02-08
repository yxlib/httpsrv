// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/yxlib/server"
	"github.com/yxlib/yx"
)

var (
	ErrNotSupportMethod = errors.New("not support http method")
	ErrNotSupportOpr    = errors.New("not support this operation")
)

type TokenDecoder interface {
	// Decode the token.
	// @param token, the token.
	// @return uint64, an id which can mark a client.
	// @return error, error.
	DecodeToken(opr string, token string) (uint64, error)
}

type Reader interface {
	// Read an request.
	// @param req, a http.Request object.
	// @param info, the config of the service which match the pattern
	// @return *server.Request, a server.Request object.
	// @return error, error.
	ReadRequest(req *http.Request, info *ServiceConf) (*server.Request, error)
}

type DefaultReader struct {
	decoder TokenDecoder
	ec      *yx.ErrCatcher
	logger  *yx.Logger
}

func NewDefaultReader(decoder TokenDecoder) *DefaultReader {
	return &DefaultReader{
		decoder: decoder,
		ec:      yx.NewErrCatcher("httpsrv.DefaultReader"),
		logger:  yx.NewLogger("httpsrv.DefaultReader"),
	}
}

func (r *DefaultReader) ReadRequest(req *http.Request, info *ServiceConf) (*server.Request, error) {
	var err error = nil
	defer r.ec.DeferThrow("ReadRequest", &err)

	// raw data
	reqData, err := r.getReqData(req)
	if err != nil {
		return nil, err
	}

	r.logger.D("Request raw data: ", reqData)

	// parse query
	val, err := ParseQuery(reqData)
	if err != nil {
		return nil, err
	}

	opr := val.Get(CfgInst.OprField)
	r.logger.I("Operation: ", opr)

	cfg, ok := info.MapOpr2Cfg[opr]
	if !ok {
		err = ErrNotSupportOpr
		return nil, err
	}

	cmd := cfg.Cmd
	r.logger.I("Command: ", cmd)
	token := val.Get(CfgInst.TokenField)
	r.logger.D("Unescape token: ", token)
	connId, err := r.decoder.DecodeToken(opr, token)
	if err != nil {
		return nil, err
	}

	// build request
	request := server.NewRequest(0)
	request.Mod = info.Mod
	request.Cmd = cmd
	request.ConnId = connId

	paramsStr := val.Get(CfgInst.ParamsField)
	request.Payload = []byte(paramsStr)
	r.logger.D("Unescape data: ", paramsStr)

	snoStr := val.Get(CfgInst.SerialNoField)
	sno, err := strconv.ParseUint(snoStr, 10, 16)
	if err == nil {
		request.SerialNo = uint16(sno)
	}

	return request, nil
}

func (r *DefaultReader) getReqData(req *http.Request) (string, error) {
	reqData := ""
	if req.Method == http.MethodGet {
		reqData = req.URL.RawQuery

	} else if req.Method == http.MethodPost {
		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return "", r.ec.Throw("getReqData", err)
		}

		reqData = string(d)

	} else {
		return "", r.ec.Throw("getReqData", ErrNotSupportMethod)
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
		key, err1 := url.PathUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.PathUnescape(value)
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
