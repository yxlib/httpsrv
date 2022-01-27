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
	DecodeToken(token string) (uint64, error)
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

	r.logger.D("raw data: ", reqData)

	// parse query
	val, err := url.ParseQuery(reqData)
	if err != nil {
		return nil, err
	}

	opr := val.Get(CfgInst.OprField)
	cfg, ok := info.MapOpr2Cfg[opr]
	if !ok {
		err = ErrNotSupportOpr
		return nil, err
	}

	cmd := cfg.Cmd
	token := val.Get(CfgInst.TokenField)
	r.logger.D("unescape token: ", token)
	connId, err := r.decoder.DecodeToken(token)
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
	r.logger.D("unescape data: ", paramsStr)

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
