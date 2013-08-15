package scenario

import (
	"strings"
)

type Call struct {
	RandomWeight      float32
	Weight            float32
	URL, Method, Body string
	Type              string // rest or www or "", default it rest

	GenParam GenCall // to generate URL & Body programmically
	CallBack GenCallBack

	SePoint *Session

	count     int64 // total # of request
	totaltime int64 // total response time.
	backlog   int64
}

func (c *Call) normalize() {
	c.Method = strings.ToUpper(c.Method)
	c.Type = strings.ToUpper(c.Type)
}

type GenCall func(ps ...string) (_m, _t, _u, _b string)
type GenCallBack func(se *Session, st int, storage []byte)
