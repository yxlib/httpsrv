// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type OprConf struct {
	Cmd     uint16 `json:"cmd"`
	Req     string `json:"req"`
	Resp    string `json:"resp"`
	Handler string `json:"handler"`
}

type ServiceConf struct {
	Service    string              `json:"service"`
	Mod        uint16              `json:"mod"`
	MapOpr2Cfg map[string]*OprConf `json:"opr"`
}

type Config struct {
	IsAllowOrigin      bool                    `json:"allow_origin"`
	OprField           string                  `json:"opr_field"`
	TokenField         string                  `json:"token_field"`
	SerialNoField      string                  `json:"serial_No_field"`
	ParamsField        string                  `json:"params_field"`
	CodeField          string                  `json:"code_field"`
	ResultField        string                  `json:"result_field"`
	MapPatten2ServInfo map[string]*ServiceConf `json:"pattern"`
}

var CfgInst *Config = &Config{}

// Load config from a json file.
// @param path, path of the json file.
// @return error, error.
func LoadConf(path string) error {
	// read file
	filePtr, err := os.Open(path)
	if err != nil {
		return err
	}

	d, err := ioutil.ReadAll(filePtr)
	if err != nil {
		return err
	}

	// json unmarshal
	err = json.Unmarshal(d, CfgInst)
	return err
}
