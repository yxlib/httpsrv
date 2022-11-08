// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"errors"
	"fmt"
	"net/http"
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
	// @param pattern, the url pattern
	// @param opr, the operation
	// @param token, the token.
	// @return uint64, an id which can mark a client.
	// @return error, error.
	DecodeToken(pattern string, opr string, token string) (uint64, error)
}

type Reader interface {
	// Read an request.
	// @param req, a http.Request object.
	// @param cfg, the config of the httpsrv
	// @return *server.Request, a server.Request object.
	// @return error, error.
	ReadRequest(req *http.Request, cfg *Config) (*server.Request, error)
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

func (r *DefaultReader) ReadRequest(req *http.Request, cfg *Config) (*server.Request, error) {
	var err error = nil
	defer r.ec.DeferThrow("ReadRequest", &err)

	// raw data
	reqData, err := GetReqData(req)
	if err != nil {
		return nil, err
	}

	r.logger.D("Request Raw Data: ", reqData)

	// parse query
	val, err := ParseQuery(reqData)
	if err != nil {
		return nil, err
	}

	// proto No.
	pattern := req.URL.Path
	r.logger.I("Pattern: ", pattern)

	opr := val.Get(cfg.OprField)
	r.logger.I("Operation: ", opr)

	procName := fmt.Sprintf("%s.%s", pattern, opr)
	protoNo, ok := cfg.Server.MapProcName2ProtoNo[procName]
	if !ok {
		err = ErrNotSupportOpr
		return nil, err
	}

	// oprCfg, ok := info.MapName2Proc[opr]
	// if !ok {
	// 	err = ErrNotSupportOpr
	// 	return nil, err
	// }

	mod := server.GetMod(protoNo)
	r.logger.I("Module: ", mod)

	cmd := server.GetCmd(protoNo)
	r.logger.I("Command: ", cmd)

	// token
	token := val.Get(cfg.TokenField)
	r.logger.D("Unescape Token: ", token)

	var connId uint64 = 0
	if r.decoder != nil {
		connId, err = r.decoder.DecodeToken(pattern, opr, token)
		if err != nil {
			return nil, err
		}
	}

	r.logger.D("Connect ID: ", connId)

	// build request
	request := server.NewRequest(0)
	request.Mod = mod
	request.Cmd = cmd
	request.ConnId = connId

	paramsStr := val.Get(cfg.ParamsField)
	request.Payload = []byte(paramsStr)
	r.logger.D("Unescape Data: ", paramsStr)

	snoStr := val.Get(cfg.SerialNoField)
	sno, err := strconv.ParseUint(snoStr, 10, 16)
	if err == nil {
		request.SerialNo = uint16(sno)
	}

	return request, nil
}
