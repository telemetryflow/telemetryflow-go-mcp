// Package mcp provides Model Context Protocol types and utilities
//
// TelemetryFlow GO MCP Server - Community Enterprise Observability Platform
// Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
// Open Source Software built by Telemetri Data Indonesia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Transport defines the interface for MCP transport
type Transport interface {
	// Read reads the next message from the transport
	Read(ctx context.Context) (*Request, error)
	// Write writes a response to the transport
	Write(ctx context.Context, response *Response) error
	// WriteNotification writes a notification to the transport
	WriteNotification(ctx context.Context, notification *Notification) error
	// Close closes the transport
	Close() error
}

// StdioTransport implements Transport using stdio
type StdioTransport struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
	closed bool
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewReader(reader),
		writer: writer,
	}
}

// Read reads the next JSON-RPC request from stdin
func (t *StdioTransport) Read(ctx context.Context) (*Request, error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil, fmt.Errorf("transport closed")
	}
	t.mu.Unlock()

	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Parse JSON-RPC request
	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Validate JSON-RPC version
	if req.JSONRPC != JSONRPCVersion {
		return nil, fmt.Errorf("invalid JSON-RPC version: %s", req.JSONRPC)
	}

	return &req, nil
}

// Write writes a JSON-RPC response to stdout
func (t *StdioTransport) Write(ctx context.Context, response *Response) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport closed")
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write with newline delimiter
	if _, err := fmt.Fprintf(t.writer, "%s\n", data); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

// WriteNotification writes a JSON-RPC notification to stdout
func (t *StdioTransport) WriteNotification(ctx context.Context, notification *Notification) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport closed")
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Write with newline delimiter
	if _, err := fmt.Fprintf(t.writer, "%s\n", data); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true
	return nil
}

// MessageHandler is a function that handles MCP requests
type MessageHandler func(ctx context.Context, req *Request) (*Response, error)

// Server wraps a transport and provides message handling
type Server struct {
	transport Transport
	handler   MessageHandler
	done      chan struct{}
	wg        sync.WaitGroup
}

// NewServer creates a new MCP server
func NewServer(transport Transport, handler MessageHandler) *Server {
	return &Server{
		transport: transport,
		handler:   handler,
		done:      make(chan struct{}),
	}
}

// Serve starts serving requests
func (s *Server) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return nil
		default:
			req, err := s.transport.Read(ctx)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				// Send error response for parse errors
				if req == nil {
					errResp := NewErrorResponse(nil, NewParseError(err.Error()))
					_ = s.transport.Write(ctx, errResp)
					continue
				}
				return err
			}

			// Handle request
			s.wg.Add(1)
			go func(req *Request) {
				defer s.wg.Done()

				resp, err := s.handler(ctx, req)
				if err != nil {
					resp = NewErrorResponse(req.ID, NewInternalError(err.Error()))
				}

				if resp != nil {
					_ = s.transport.Write(ctx, resp)
				}
			}(req)
		}
	}
}

// Stop stops the server
func (s *Server) Stop() {
	close(s.done)
	s.wg.Wait()
	_ = s.transport.Close()
}

// SendNotification sends a notification
func (s *Server) SendNotification(ctx context.Context, method string, params interface{}) error {
	notification, err := NewNotification(method, params)
	if err != nil {
		return err
	}
	return s.transport.WriteNotification(ctx, notification)
}
