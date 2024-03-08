package gsse

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"
)

func (c *Client) SendMessage(data string) {
	c.emit(noEvent, data, noId)
}

func (c *Client) SendMessageWithId(data, id string) {
	c.emit(noEvent, data, id)
}

func (c *Client) SendEvent(event, data string) {
	c.emit(event, data, noId)
}

func (c *Client) SendEventWithId(event, data, id string) {
	c.emit(event, data, id)
}

func (c *Client) emit(event, data, id string) {
	// default event: message
	if event != noEvent {
		c.Response.Writeln("event:", event)
	} else {
		c.Response.Writeln("event:", message)
	}
	c.Response.Writeln("data:", data)
	// default id: no id
	if id != noId {
		c.Response.Writeln("id:", id)
	}
	c.Response.Writeln()
	c.Response.Flush()
}

func (c *Client) SendComment(comment string) {
	c.Response.Writeln(":", comment)
	c.Response.Writeln()
	c.Response.Flush()
}

// Close closes the connection
func (c *Client) Close() {
	c.cancel()
}

// Terminated returns true if the connection has been closed
func (c *Client) Terminated() bool {
	return c.Context.Err() != nil
}

// OnClose callback which runs when a client closes its connection
func (c *Client) OnClose(fn func(*Client)) {
	c.onClose = fn
}

// KeepAlive keeps the connection alive, if you need to use the client outside the handler
func (c *Client) KeepAlive() {
	c.keepAlive = true
}

func newClient(request *ghttp.Request) *Client {
	ctx, cancel := context.WithCancel(request.Context())
	request.SetCtx(ctx)
	response := request.Response
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	return &Client{
		Context:  ctx,
		Request:  request,
		Response: response,
		Server:   request.Server,

		cancel:    cancel,
		onClose:   nil,
		keepAlive: false,
	}
}
