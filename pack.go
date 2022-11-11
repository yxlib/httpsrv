package httpsrv

type Request struct {
	Token    string
	Pattern  string
	Opr      string
	SerialNo uint16
	Params   string
}

type Response struct {
	Opr      string
	SerialNo uint16
	Code     int32
	Result   string
}
