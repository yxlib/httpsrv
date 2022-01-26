// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"errors"
	"net/http"

	"github.com/yxlib/server"
	"github.com/yxlib/yx"
)

var (
	ErrUnknownPattern = errors.New("unknown path pattern")
)

type HttpServer struct {
	*server.Server
	reader       Reader
	writer       Writer
	bAllowOrigin bool
	ec           *yx.ErrCatcher
	logger       *yx.Logger
}

func NewHttpServer(r Reader, w Writer, bAllowOrigin bool) *HttpServer {
	return &HttpServer{
		Server:       server.NewServer("HttpServer", nil),
		reader:       r,
		writer:       w,
		bAllowOrigin: bAllowOrigin,
		ec:           yx.NewErrCatcher("HttpServer"),
		logger:       yx.NewLogger("HttpServer"),
	}
}

// Bind pattern to service.
// @param pattern, the http pattern.
// @param mod, the module of the service.
// @param srv, the service.
func (s *HttpServer) Bind(pattern string, mod uint16, srv server.Service) {
	s.AddService(srv, mod)
	http.HandleFunc(pattern, s.handleFunc)
}

// Start the http server.
// @param addr, the http address.
// @return error, error.
func (s *HttpServer) Start(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (s *HttpServer) handleFunc(w http.ResponseWriter, req *http.Request) {
	if !s.handleOrign(w, req) {
		return
	}

	var err error = nil
	respCode := int(0)
	respResult := ""

	// write response
	defer func() {
		s.writer.WriteResponse(w, respCode, respResult, err)
	}()

	defer s.ec.Catch("handleFunc", &err)

	// create request
	request, code, err := s.createRequest(req)
	if err != nil {
		respCode = code
		return
	}

	// handle
	response := server.NewResponse(request)
	s.HandleHttpRequest(request, response)

	// result
	respCode = int(response.Code)
	respResult = string(response.Payload)
}

func (s *HttpServer) handleOrign(w http.ResponseWriter, req *http.Request) bool {
	if !s.bAllowOrigin {
		return req.Method != "OPTIONS"
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}

	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return false
	}

	return true
}

func (s *HttpServer) createRequest(req *http.Request) (*server.Request, int, error) {
	var err error = nil
	defer s.ec.DeferThrow("createRequest", &err)

	pattern := req.URL.Path
	info, ok := CfgInst.MapPatten2ServInfo[pattern]
	if !ok {
		return nil, RESP_CODE_UNKNOWN_PATTERN, ErrUnknownPattern
	}

	request, err := s.reader.ReadRequest(req, info)
	if err != nil {
		return nil, RESP_CODE_DECODE_FAILED, err
	}

	return request, server.RESP_CODE_SUCCESS, nil
}
