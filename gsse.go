package gsse

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Client struct {
	Context  context.Context
	Request  *ghttp.Request
	Response *ghttp.Response
	Server   *ghttp.Server

	cancel    context.CancelFunc
	onClose   func(*Client)
	keepAlive bool
}

const (
	noEvent = ""
	message = "message"

	noId = ""

	emptyComment = ""
)
