// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"os"
	"strings"

	"github.com/yxlib/yx"
)

func GenServiceRegisterFile(cfgPath string, regFilePath string, regPackName string) {
	yx.LoadJsonConf(CfgInst, cfgPath, nil)
	GenRegisterFileByCfg(CfgInst, regFilePath, regPackName)
}

// Generate the service register file.
// @param cfgPath, the config path.
// @param regFilePath, the output register file.
// @param regPackName, the package name of the file.
func GenRegisterFileByCfg(srvCfg *Config, regFilePath string, regPackName string) {
	f, err := os.OpenFile(regFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	defer f.Close()

	f.WriteString("// This File auto generate by tool.\n")
	f.WriteString("// Please do not modify.\n")
	f.WriteString("// See httpsrv.GenRegisterFileByCfg().\n\n")
	f.WriteString("package " + regPackName + "\n\n")
	f.WriteString("import (\n")

	packSet := yx.NewSet(yx.SET_TYPE_OBJ)
	packSet.Add("github.com/yxlib/server")

	for _, servCfg := range srvCfg.MapPatten2ServInfo {
		servStr := servCfg.Service
		idx := strings.LastIndex(servStr, ".")
		packSet.Add(servStr[:idx])
		for _, cfg := range servCfg.MapOpr2Cfg {
			reqStr := cfg.Req
			idx := strings.LastIndex(reqStr, ".")
			packSet.Add(reqStr[:idx])

			respStr := cfg.Req
			idx = strings.LastIndex(respStr, ".")
			packSet.Add(respStr[:idx])
		}
	}

	elements := packSet.GetElements()
	for _, packName := range elements {
		f.WriteString("    \"" + packName.(string) + "\"\n")
	}

	f.WriteString(")\n\n")

	f.WriteString("// Auto generate by tool.\n")
	f.WriteString("func RegisterServices() {\n")

	for pattern, servCfg := range srvCfg.MapPatten2ServInfo {
		f.WriteString("    //===============================\n")
		f.WriteString("    //        " + pattern + "\n")
		f.WriteString("    //===============================\n")

		servStr := servCfg.Service
		idx := strings.LastIndex(servStr, "/")
		if idx >= 0 {
			servStr = servStr[idx+1:]
		}
		idx = strings.Index(servStr, ".")
		packName := servStr[:idx]
		servStr = servStr[idx+1:]
		f.WriteString("    server.ServiceBinder.BindService(" + packName + ".New" + servStr + "())\n")
		for opr, cfg := range servCfg.MapOpr2Cfg {
			f.WriteString("    // " + opr + "\n")

			reqStr := cfg.Req
			idx := strings.LastIndex(reqStr, "/")
			if idx >= 0 {
				reqStr = reqStr[idx+1:]
			}
			f.WriteString("    server.ProtoBinder.RegisterProto(&" + reqStr + "{})\n")

			respStr := cfg.Resp
			idx = strings.LastIndex(respStr, "/")
			if idx >= 0 {
				respStr = respStr[idx+1:]
			}
			f.WriteString("    server.ProtoBinder.RegisterProto(&" + respStr + "{})\n")
		}

		f.WriteString("\n")
	}

	f.WriteString("}")

}
