// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

type Config struct {
	OprField      string `json:"opr_field"`
	TokenField    string `json:"token_field"`
	SerialNoField string `json:"serial_No_field"`
	ParamsField   string `json:"params_field"`
	CodeField     string `json:"code_field"`
	ResultField   string `json:"result_field"`
}
