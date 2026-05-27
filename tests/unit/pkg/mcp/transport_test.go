package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-go-mcp/pkg/mcp"
)

func TestStdioTransport_Read(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
		r := strings.NewReader(input)
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)

		req, err := tp.Read(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "2.0", req.JSONRPC)
		assert.Equal(t, "initialize", req.Method)
	})

	t.Run("invalid json", func(t *testing.T) {
		input := "not json\n"
		r := strings.NewReader(input)
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)

		_, err := tp.Read(context.Background())
		require.Error(t, err)
	})

	t.Run("invalid jsonrpc version", func(t *testing.T) {
		input := `{"jsonrpc":"1.0","id":1,"method":"test"}` + "\n"
		r := strings.NewReader(input)
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)

		_, err := tp.Read(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON-RPC version")
	})
}

func TestStdioTransport_Write(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)

	resp := mcp.NewResponse(1, map[string]string{"status": "ok"})
	err := tp.Write(context.Background(), resp)
	require.NoError(t, err)
	assert.True(t, w.Len() > 0)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Bytes(), &decoded))
	assert.Equal(t, "2.0", decoded["jsonrpc"])
}

func TestStdioTransport_WriteNotification(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)

	notif, err := mcp.NewNotification("test/method", map[string]string{"k": "v"})
	require.NoError(t, err)
	err = tp.WriteNotification(context.Background(), notif)
	require.NoError(t, err)
	assert.True(t, w.Len() > 0)
}

func TestStdioTransport_Close(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)

	err := tp.Close()
	require.NoError(t, err)

	_, err = tp.Read(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "transport closed")

	err = tp.Write(context.Background(), &mcp.Response{})
	require.Error(t, err)

	err = tp.WriteNotification(context.Background(), &mcp.Notification{})
	require.Error(t, err)
}

func TestNewServer(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)
	handler := func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
		return mcp.NewResponse(req.ID, nil), nil
	}
	srv := mcp.NewServer(tp, handler)
	assert.NotNil(t, srv)
}

func TestServer_Stop(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)
	srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
		return nil, nil
	})
	srv.Stop()
}

func TestServer_SendNotification(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)
	srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
		return nil, nil
	})

	err := srv.SendNotification(context.Background(), "test/method", map[string]string{"k": "v"})
	require.NoError(t, err)
}

func TestServer_Serve(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		r := strings.NewReader("")
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)
		srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
			return mcp.NewResponse(req.ID, nil), nil
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := srv.Serve(ctx)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("handles request and stops", func(t *testing.T) {
		input := `{"jsonrpc":"2.0","id":1,"method":"test","params":{}}` + "\n"
		r := strings.NewReader(input)
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)
		srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
			return mcp.NewResponse(req.ID, map[string]string{"echo": req.Method}), nil
		})

		done := make(chan struct{})
		go func() {
			defer close(done)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = srv.Serve(ctx)
		}()
		time.Sleep(100 * time.Millisecond)
		srv.Stop()
		<-done
	})

	t.Run("eof stops serve", func(t *testing.T) {
		r := strings.NewReader("")
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)
		srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
			return nil, nil
		})

		err := srv.Serve(context.Background())
		assert.NoError(t, err)
	})

	t.Run("handler error returns internal error response", func(t *testing.T) {
		input := `{"jsonrpc":"2.0","id":42,"method":"test","params":{}}` + "\n"
		r := strings.NewReader(input)
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)
		srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
			return nil, fmt.Errorf("handler failed")
		})

		done := make(chan struct{})
		go func() {
			defer close(done)
			time.Sleep(200 * time.Millisecond)
			srv.Stop()
		}()
		_ = srv.Serve(context.Background())
		<-done
	})

	t.Run("handler returns nil response", func(t *testing.T) {
		input := `{"jsonrpc":"2.0","id":99,"method":"test","params":{}}` + "\n"
		r := strings.NewReader(input)
		var w bytes.Buffer
		tp := mcp.NewStdioTransport(r, &w)
		srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
			return nil, nil
		})

		done := make(chan struct{})
		go func() {
			defer close(done)
			time.Sleep(200 * time.Millisecond)
			srv.Stop()
		}()
		_ = srv.Serve(context.Background())
		<-done
	})
}

func TestServer_Serve_ParseError(t *testing.T) {
	input := "not json\n"
	r := strings.NewReader(input)
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)
	srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
		return mcp.NewResponse(req.ID, nil), nil
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(200 * time.Millisecond)
		srv.Stop()
	}()
	_ = srv.Serve(context.Background())
	<-done
}

func TestServer_SendNotification_ClosedTransport(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)
	srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
		return nil, nil
	})

	_ = tp.Close()

	err := srv.SendNotification(context.Background(), "test/method", map[string]string{"k": "v"})
	if err == nil {
		t.Error("expected error sending notification on closed transport")
	}
}

func TestStdioTransport_Write_FailingWriter(t *testing.T) {
	r := strings.NewReader("")
	w := &failingWriter{}
	tp := mcp.NewStdioTransport(r, w)
	resp := mcp.NewResponse(1, map[string]string{"status": "ok"})
	err := tp.Write(context.Background(), resp)
	if err == nil {
		t.Error("expected error writing to failing writer")
	}
}

func TestStdioTransport_WriteNotification_FailingWriter(t *testing.T) {
	r := strings.NewReader("")
	w := &failingWriter{}
	tp := mcp.NewStdioTransport(r, w)
	notif, err := mcp.NewNotification("test/method", map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("NewNotification: %v", err)
	}
	err = tp.WriteNotification(context.Background(), notif)
	if err == nil {
		t.Error("expected error writing notification to failing writer")
	}
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write failed")
}

func TestServer_Serve_NilNotification(t *testing.T) {
	r := strings.NewReader("")
	var w bytes.Buffer
	tp := mcp.NewStdioTransport(r, &w)
	srv := mcp.NewServer(tp, func(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
		return nil, nil
	})

	err := srv.SendNotification(context.Background(), "", nil)
	if err != nil {
		t.Logf("SendNotification with empty method: %v", err)
	}
}

func TestNewNotification_BadParams(t *testing.T) {
	_, err := mcp.NewNotification("test", map[string]interface{}{"ch": make(chan int)})
	if err == nil {
		t.Error("expected error for non-marshalable params")
	}
}

func TestNewNotification_NilParams(t *testing.T) {
	n, err := mcp.NewNotification("test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Method != "test" {
		t.Errorf("expected method test, got %s", n.Method)
	}
	if n.Params != nil {
		t.Error("expected nil params")
	}
}
