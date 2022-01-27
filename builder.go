// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpsrv

import (
	"reflect"

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
	for pattern, servCfg := range cfg.MapPatten2ServInfo {
		b.parsePatternCfg(srv, pattern, servCfg)
	}
}

func (b *builder) parsePatternCfg(srv *Server, pattern string, servCfg *ServiceConf) {
	s, ok := server.ServiceBinder.GetService(servCfg.Service)
	if !ok {
		b.logger.W("Not support pattern ", pattern)
		return
	}

	srv.Bind(pattern, servCfg.Mod, s)
	b.parseOprCfg(s, servCfg)
}

func (b *builder) parseOprCfg(s server.Service, servCfg *ServiceConf) {
	v := reflect.ValueOf(s)

	for opr, cfg := range servCfg.MapOpr2Cfg {
		// proto
		err := server.ProtoBinder.BindProto(servCfg.Mod, cfg.Cmd, cfg.Req, cfg.Resp)
		if err != nil {
			b.logger.W("not support operation ", opr)
		}

		// processor
		m := v.MethodByName(cfg.Handler)
		err = s.AddReflectProcessor(m, cfg.Cmd)
		if err != nil {
			b.logger.E("AddReflectProcessor err: ", err)
			b.logger.W("not support operation ", opr)
			continue
		}
	}
}
