// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"github.com/yxlib/server"
	"github.com/yxlib/yx"
)

type builder struct {
	logger *yx.Logger
}

var Builder = &builder{
	logger: yx.NewLogger("httpsrv.Builder"),
}

// build server.
// @param srv, dest http server.
// @param cfg, the server config.
func (b *builder) Build(srv *Server, cfg *Config) {
	mod := uint16(1)
	for _, service := range cfg.Server.Services {
		service.Mod = mod
		mod++

		// s, ok := server.ServiceBinder.GetService(service.Service)
		// if !ok {
		// 	b.logger.W("Not support pattern ", service.Name)
		// 	return
		// }
		// srv.Bind(service.Name, service.Mod, s)

		cmd := uint16(1)
		for _, proc := range service.Processors {
			proc.Cmd = cmd
			cmd++
		}
	}

	server.Builder.Build(srv, cfg.Server)
	// for pattern, servCfg := range cfg.MapPatten2ServInfo {
	// 	servCfg.Mod = mod
	// 	mod++
	// 	b.parsePatternCfg(srv, pattern, servCfg)
	// }
}

// func (b *builder) parsePatternCfg(srv *Server, pattern string, servCfg *ServiceConf) {
// 	s, ok := server.ServiceBinder.GetService(servCfg.Service)
// 	if !ok {
// 		b.logger.W("Not support pattern ", pattern)
// 		return
// 	}

// 	srv.Bind(pattern, servCfg.Mod, s)
// 	b.parseOprCfg(s, servCfg)
// }

// func (b *builder) parseOprCfg(s server.Service, servCfg *ServiceConf) {
// 	v := reflect.ValueOf(s)

// 	cmd := uint16(1)
// 	for opr, cfg := range servCfg.MapOpr2Cfg {
// 		cfg.Cmd = cmd
// 		cmd++

// 		// proto
// 		err := server.ProtoBinder.BindProto(servCfg.Mod, cfg.Cmd, cfg.Req, cfg.Resp)
// 		if err != nil {
// 			b.logger.W("not support operation ", opr)
// 			continue
// 		}

// 		// processor
// 		m := v.MethodByName(cfg.Handler)
// 		err = s.AddReflectProcessor(m, cfg.Cmd)
// 		if err != nil {
// 			b.logger.E("AddReflectProcessor err: ", err)
// 			b.logger.W("not support operation ", opr)
// 			continue
// 		}
// 	}
// }
