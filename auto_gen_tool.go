// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"os"
	"strings"

	"github.com/yxlib/server"
	"github.com/yxlib/yx"
)

func GenServiceRegisterFile(cfgPath string, regFilePath string, regPackName string) {
	cfg := &Config{}
	yx.LoadJsonConf(cfg, cfgPath, nil)
	GenRegisterFileByCfg(cfg, regFilePath, regPackName)
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

	writePackage(f, regPackName)
	writeImport(srvCfg.Server.MapName2Service, regPackName, f)

	f.WriteString("// Auto generate by tool.\n")
	f.WriteString("func RegisterServices() {\n")

	for pattern, servCfg := range srvCfg.Server.MapName2Service {
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
		for opr, cfg := range servCfg.MapName2Proc {
			// if cfg.Req == "" && cfg.Resp == "" {
			// 	continue
			// }

			f.WriteString("    // " + opr + "\n")

			if cfg.Req != "" {
				reqStr := getFilePackageClassName(cfg.Req, regPackName)
				f.WriteString("    server.ProtoBinder.RegisterProto(&" + reqStr + "{})\n")
			}

			if cfg.Resp != "" {
				respStr := getFilePackageClassName(cfg.Resp, regPackName)
				f.WriteString("    server.ProtoBinder.RegisterProto(&" + respStr + "{})\n")
			}
		}

		f.WriteString("\n")
	}

	f.WriteString("}")

}

func getFilePackageClassName(classReflectName string, regPackName string) string {
	fullPackName := yx.GetFullPackageName(classReflectName)
	filePackName := yx.GetFilePackageName(fullPackName)

	if filePackName == regPackName {
		return yx.GetClassName(classReflectName)
	}

	return yx.GetFilePackageClassName(classReflectName)
}

func writePackage(f *os.File, regPackName string) {
	f.WriteString("// This File auto generate by tool.\n")
	f.WriteString("// Please do not modify.\n")
	f.WriteString("// See httpsrv.GenRegisterFileByCfg().\n\n")
	f.WriteString("package " + regPackName + "\n\n")
}

func writeImport(mapPatten2ServInfo map[string]*server.ServiceConf, regPackName string, f *os.File) {
	f.WriteString("import (\n")

	packSet := yx.NewSet(yx.SET_TYPE_OBJ)
	packSet.Add("github.com/yxlib/server")

	for _, servCfg := range mapPatten2ServInfo {
		fullPackName := yx.GetFullPackageName(servCfg.Service)
		if fullPackName != "" {
			packSet.Add(fullPackName)
		}

		for _, cfg := range servCfg.MapName2Proc {
			addProtoPackage(cfg.Req, regPackName, packSet)
			addProtoPackage(cfg.Resp, regPackName, packSet)
		}
	}

	elements := packSet.GetElements()
	for _, packName := range elements {
		f.WriteString("    \"" + packName.(string) + "\"\n")
	}

	f.WriteString(")\n\n")
}

func addProtoPackage(protoCfg string, regPackName string, packSet *yx.Set) {
	if protoCfg == "" {
		return
	}

	fullPackName := yx.GetFullPackageName(protoCfg)
	filePackName := yx.GetFilePackageName(fullPackName)
	if filePackName != regPackName {
		packSet.Add(fullPackName)
	}
}
