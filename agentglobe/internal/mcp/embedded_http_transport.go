package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/metoro-io/mcp-golang/transport"
)

// embeddedHTTPTransport is a stateless HTTP MCP transport that does not listen on its own port.
// It implements [transport.Transport] with a no-op [embeddedHTTPTransport.Start] so [mcp_golang.Server.Serve]
// can wire the protocol, then the same value is mounted as [http.Handler] on agentglobe at /mcp.
//
// Adapted from github.com/metoro-io/mcp-golang/transport/http (baseTransport + handleRequest).
type embeddedHTTPTransport struct {
	messageHandler func(ctx context.Context, message *transport.BaseJsonRpcMessage)
	errorHandler   func(error)
	closeHandler   func()
	mu             sync.RWMutex
	responseMap    map[int64]chan *transport.BaseJsonRpcMessage
}

func newEmbeddedHTTPTransport() *embeddedHTTPTransport {
	return &embeddedHTTPTransport{
		responseMap: make(map[int64]chan *transport.BaseJsonRpcMessage),
	}
}

func (t *embeddedHTTPTransport) getResponseChannel(key int64) chan *transport.BaseJsonRpcMessage {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.responseMap[key]
}

func (t *embeddedHTTPTransport) reserveResponseChannel() (int64, chan *transport.BaseJsonRpcMessage) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var key int64
	for key < 1000000 {
		if _, ok := t.responseMap[key]; !ok {
			break
		}
		key++
	}

	ch := make(chan *transport.BaseJsonRpcMessage)
	t.responseMap[key] = ch
	return key, ch
}

func (t *embeddedHTTPTransport) deleteResponseChannel(key int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.responseMap, key)
}

// Start implements [transport.Transport]; the real listener is agentglobe’s HTTP server.
func (t *embeddedHTTPTransport) Start(context.Context) error {
	return nil
}

// Send implements [transport.Transport].
func (t *embeddedHTTPTransport) Send(ctx context.Context, message *transport.BaseJsonRpcMessage) error {
	_ = ctx
	key := message.JsonRpcResponse.Id
	ch := t.getResponseChannel(int64(key))
	if ch == nil {
		return fmt.Errorf("no response channel found for key: %d", key)
	}
	ch <- message
	return nil
}

// Close implements [transport.Transport].
func (t *embeddedHTTPTransport) Close() error {
	if t.closeHandler != nil {
		t.closeHandler()
	}
	return nil
}

// SetCloseHandler implements [transport.Transport].
func (t *embeddedHTTPTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeHandler = handler
}

// SetErrorHandler implements [transport.Transport].
func (t *embeddedHTTPTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = handler
}

// SetMessageHandler implements [transport.Transport].
func (t *embeddedHTTPTransport) SetMessageHandler(handler func(ctx context.Context, message *transport.BaseJsonRpcMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

func (t *embeddedHTTPTransport) handleMessage(ctx context.Context, body []byte) (*transport.BaseJsonRpcMessage, error) {
	key, responseCh := t.reserveResponseChannel()

	var prevID *transport.RequestId
	deserialized := false
	var request transport.BaseJSONRPCRequest
	if err := json.Unmarshal(body, &request); err == nil {
		deserialized = true
		id := request.Id
		prevID = &id
		request.Id = transport.RequestId(key)
		t.mu.RLock()
		h := t.messageHandler
		t.mu.RUnlock()
		if h != nil {
			h(ctx, transport.NewBaseMessageRequest(&request))
		}
	}

	var notification transport.BaseJSONRPCNotification
	if !deserialized {
		if err := json.Unmarshal(body, &notification); err == nil {
			deserialized = true
			t.mu.RLock()
			h := t.messageHandler
			t.mu.RUnlock()
			if h != nil {
				h(ctx, transport.NewBaseMessageNotification(&notification))
			}
		}
	}

	var response transport.BaseJSONRPCResponse
	if !deserialized {
		if err := json.Unmarshal(body, &response); err == nil {
			deserialized = true
			t.mu.RLock()
			h := t.messageHandler
			t.mu.RUnlock()
			if h != nil {
				h(ctx, transport.NewBaseMessageResponse(&response))
			}
		}
	}

	var errorResponse transport.BaseJSONRPCError
	if !deserialized {
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			deserialized = true
			t.mu.RLock()
			h := t.messageHandler
			t.mu.RUnlock()
			if h != nil {
				h(ctx, transport.NewBaseMessageError(&errorResponse))
			}
		}
	}

	out := <-responseCh
	t.deleteResponseChannel(key)
	if prevID != nil {
		out.JsonRpcResponse.Id = *prevID
	}
	return out, nil
}

func (t *embeddedHTTPTransport) readBody(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		if t.errorHandler != nil {
			t.errorHandler(fmt.Errorf("failed to read request body: %w", err))
		}
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	return body, nil
}

// ServeHTTP handles MCP JSON-RPC over HTTP POST (same contract as standalone MCP_HTTP_PATH).
func (t *embeddedHTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}
	body, err := t.readBody(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := t.handleMessage(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		if t.errorHandler != nil {
			t.errorHandler(fmt.Errorf("failed to marshal response: %w", err))
		}
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}
