package gsse_test

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/CharLemAznable/gsse"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/guid"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_SendMessage(t *testing.T) {
	s := g.Server(guid.S())
	s.BindHandler("/sse", gsse.Handle(func(client *gsse.Client) {
		client.SendMessage("send message")
	}))
	s.SetDumpRouterMap(false)
	_ = s.Start()
	defer func() { _ = s.Shutdown() }()

	time.Sleep(100 * time.Millisecond)
	gtest.C(t, func(t *gtest.T) {
		prefix := fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())
		client := g.Client()
		client.SetPrefix(prefix)

		t.Assert(client.GetContent(gctx.New(), "/sse"),
			"event:message\ndata:send message\n\n")
	})
}

func Test_SendMessageWithId(t *testing.T) {
	ch := make(chan bool, 1)
	s := g.Server(guid.S())
	s.BindHandler("/sse", gsse.Handle(func(client *gsse.Client) {
		client.OnClose(func(client *gsse.Client) {
			ch <- client.Terminated()
		})
		client.SendMessageWithId("send message with id", "1")
	}))
	s.SetDumpRouterMap(false)
	_ = s.Start()
	defer func() { _ = s.Shutdown() }()

	time.Sleep(100 * time.Millisecond)
	gtest.C(t, func(t *gtest.T) {
		prefix := fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())
		client := g.Client()
		client.SetPrefix(prefix)

		t.Assert(client.GetContent(gctx.New(), "/sse"),
			"event:message\ndata:send message with id\nid:1\n\n")

		select {
		case value := <-ch:
			t.AssertEQ(value, true)
		}
	})
}

type eventSource struct {
	Lines chan string
	Done  chan error
}

func newEventSource(url string) *eventSource {
	es := &eventSource{
		Lines: make(chan string),
		Done:  make(chan error),
	}

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			es.Done <- err
			return
		}
		defer func() { _ = resp.Body.Close() }()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			es.Lines <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			es.Done <- err
		} else {
			es.Done <- io.EOF
		}
	}()

	return es
}

func Test_SendEvent(t *testing.T) {
	clientCh := make(chan *gsse.Client, 1)
	s := g.Server(guid.S())
	s.BindHandler("/sse", gsse.Handle(func(client *gsse.Client) {
		client.KeepAlive()
		clientCh <- client
	}))
	s.SetDumpRouterMap(false)
	_ = s.Start()
	defer func() { _ = s.Shutdown() }()

	time.Sleep(100 * time.Millisecond)
	gtest.C(t, func(t *gtest.T) {
		prefix := fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())
		es := newEventSource(prefix + "/sse")

		go func() {
			sseClient := <-clientCh
			sseClient.SendEvent("test", "send event")
			sseClient.Close()
		}()
		finish := make(chan interface{}, 1)
		go func() {
			var buffer bytes.Buffer
			for {
				select {
				case line := <-es.Lines:
					buffer.WriteString(line)
					//buffer.WriteString("\n")
				case <-es.Done:
					t.Assert(gstr.TrimStr(buffer.String(), ":"),
						"event:testdata:send event")
					finish <- ""
					return
				}
			}
		}()
		<-finish
	})
}

func Test_SendEventWithId(t *testing.T) {
	clientCh := make(chan *gsse.Client, 1)
	s := g.Server(guid.S())
	s.BindHandler("/sse", gsse.Handle(func(client *gsse.Client) {
		client.KeepAlive()
		clientCh <- client
	}))
	s.SetDumpRouterMap(false)
	_ = s.Start()
	defer func() { _ = s.Shutdown() }()

	time.Sleep(100 * time.Millisecond)
	gtest.C(t, func(t *gtest.T) {
		prefix := fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())
		es := newEventSource(prefix + "/sse")

		go func() {
			sseClient := <-clientCh
			sseClient.SendEventWithId("test", "send event", "2")
			time.Sleep(7 * time.Second)
			sseClient.Close()
		}()
		finish := make(chan interface{}, 1)
		go func() {
			var buffer bytes.Buffer
			for {
				select {
				case line := <-es.Lines:
					buffer.WriteString(line)
					//buffer.WriteString("\n")
				case <-es.Done:
					t.Assert(gstr.TrimStr(buffer.String(), ":"),
						"event:testdata:send eventid:2")
					finish <- ""
					return
				}
			}
		}()
		<-finish
	})
}
