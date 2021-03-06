// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/yxlib/yx"
)

type Writer interface {
	// Write a response.
	// @param writer, the http.ResponseWriter object.
	// @param respCode, the response code.
	// @param respResult, the response result data.
	// @param error, error.
	WriteResponse(writer http.ResponseWriter, respCode int, respResult string, err error, cfg *Config)
}

type DefaultWriter struct {
	ec     *yx.ErrCatcher
	logger *yx.Logger
}

func NewDefaultWriter() *DefaultWriter {
	return &DefaultWriter{
		ec:     yx.NewErrCatcher("httpsrv.DefaultWriter"),
		logger: yx.NewLogger("httpsrv.DefaultWriter"),
	}
}

func (w *DefaultWriter) WriteResponse(writer http.ResponseWriter, respCode int, respResult string, err error, cfg *Config) {
	writer.Header().Set("content-type", "application/x-www-form-urlencoded")

	if err != nil {
		respResult = "\"" + err.Error() + "\""
	}

	respResult = url.PathEscape(respResult)
	respData := cfg.CodeField + "=" + strconv.Itoa(respCode) + "&" + cfg.ResultField + "=" + string(respResult)
	w.logger.D("Response raw data: ", respData)
	w.logger.D()

	_, errWrite := writer.Write([]byte(respData))
	w.ec.Catch("WriteResponse", &errWrite)
}
