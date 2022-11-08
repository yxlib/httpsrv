package httpsrv

import (
	"strconv"
	"testing"

	"github.com/yxlib/server"
	"github.com/yxlib/yx"
)

type Cmd1Request struct {
	Addr string `json:"a"`
	Port uint16 `json:"p"`
}

func (r *Cmd1Request) Reset() {
	r.Port = 0
}

type Cmd1Response struct {
	No uint16 `json:"n"`
}

func (r *Cmd1Response) Reset() {
	r.No = 0
}

type Cmd2Request struct {
	Addr string `json:"a"`
	Port uint16 `json:"p"`
}

func (r *Cmd2Request) Reset() {
	r.Port = 0
}

type Cmd2Response struct {
	No uint16 `json:"n"`
}

func (r *Cmd2Response) Reset() {
	r.No = 0
}

type ExampleService struct {
	*server.BaseService
	logger *yx.Logger
}

func NewExampleService() *ExampleService {
	s := &ExampleService{
		BaseService: server.NewBaseService("ExampleService"),
		logger:      yx.NewLogger("ExampleService"),
	}

	return s
}

func (s *ExampleService) OnHandleCmd1(req *server.Request, resp *server.Response) (int32, error) {
	s.logger.D("onHandleCmd1")

	reqData := req.ExtData.(*Cmd1Request)
	s.logger.D("request info: ", reqData.Addr, ":", reqData.Port)

	respData := resp.ExtData.(*Cmd1Response)
	respData.No = 31
	return 0, nil
}

func (s *ExampleService) OnHandleCmd2(req *server.Request, resp *server.Response) (int32, error) {
	s.logger.D("onHandleCmd2")

	reqData := req.ExtData.(*Cmd2Request)
	s.logger.D("request info: ", reqData.Addr, ":", reqData.Port)

	respData := resp.ExtData.(*Cmd2Response)
	respData.No = 32
	return 0, nil
}

func registerProtos() {
	server.ProtoBinder.RegisterProto(&Cmd1Request{})
	server.ProtoBinder.RegisterProto(&Cmd1Response{})

	server.ProtoBinder.RegisterProto(&Cmd2Request{})
	server.ProtoBinder.RegisterProto(&Cmd2Response{})
}

func registerServices() {
	server.ServiceBinder.BindService(NewExampleService())
}

type TestTokenDecoder struct {
}

func (d *TestTokenDecoder) DecodeToken(pattern string, opr string, token string) (uint64, error) {
	connId, err := strconv.ParseUint(token, 10, 64)
	if err != nil {
		return 0, err
	}

	return connId, nil
}

func TestHttpSrv(t *testing.T) {
	registerProtos()
	registerServices()

	cfg := &Config{}
	err := yx.LoadJsonConf(cfg, "cfg_template.json", nil)
	if err != nil {
		panic(err)
	}

	d := &TestTokenDecoder{}
	r := NewDefaultReader(d)
	w := NewDefaultWriter()
	s := NewServer(r, w, cfg)
	s.AddGlobalInterceptor(&server.JsonInterceptor{})
	server.Builder.Build(s, cfg.Server)

	logger := yx.NewLogger("TestHttpSrv")
	logger.I("###### http server start ######")

	s.Listen(":8080")

	logger.I("###### http server stop ######")
}

func TestGenRegFile(t *testing.T) {
	server.GenServiceRegisterFile("cfg_template.json", "serv_reg.goxx", "httpsrv")
}
