// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/yxlib/yx"
)

var (
	ErrUnknownPattern = errors.New("unknown path pattern")
)

//========================
//      HttpListener
//========================
type HttpListener interface {
	OnHttpReadPack(w http.ResponseWriter, req *http.Request) error
}

type Server struct {
	listener     HttpListener
	bAllowOrigin bool
	httpSrv      *http.Server
	evtShutdown  *yx.Event

	ec     *yx.ErrCatcher
	logger *yx.Logger
}

func NewServer(l HttpListener) *Server {
	return &Server{
		listener:     l,
		bAllowOrigin: false,
		httpSrv:      nil,
		evtShutdown:  yx.NewEvent(),

		ec:     yx.NewErrCatcher("httpsrv.Server"),
		logger: yx.NewLogger("httpsrv.Server"),
	}
}

// Start the http server.
// @param addr, the http address.
// @return error, error.
func (s *Server) Listen(addr string) error {
	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: s,
	}

	err := s.httpSrv.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	s.evtShutdown.Wait()
	return nil
}

func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := s.httpSrv.Shutdown(ctx)
	if err != nil {
		return err
	}

	s.evtShutdown.Send()
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !s.handleOrign(w, req) {
		return
	}

	err := s.listener.OnHttpReadPack(w, req)
	s.ec.Catch("ServeHTTP", &err)
}

func (s *Server) handleOrign(w http.ResponseWriter, req *http.Request) bool {
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
