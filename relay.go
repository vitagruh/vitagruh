// === FILE: relay.go ===
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Logger wraps standard log with level filtering
type Logger struct {
	level  LogLevel
	debug  *log.Logger
	info   *log.Logger
	err    *log.Logger
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level: level,
		debug: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lmicroseconds),
		info:  log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lmicroseconds),
		err:   log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmicroseconds),
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level == LogLevelDebug {
		l.debug.Printf(format, args...)
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.level == LogLevelDebug || l.level == LogLevelInfo {
		l.info.Printf(format, args...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.err.Printf(format, args...)
}

// BackendResult holds the result from a backend ICAP server
type BackendResult struct {
	Name     string
	Response *ICAPResponse
	Error    error
}

// Relay represents the ICAP relay server
type Relay struct {
	config   *Config
	logger   *Logger
	listener net.Listener
	maxConns chan struct{} // Semaphore for connection limiting
}

// NewRelay creates a new ICAP relay instance
func NewRelay(cfg *Config) (*Relay, error) {
	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	relay := &Relay{
		config:   cfg,
		logger:   NewLogger(cfg.LogLevel),
		listener: ln,
		maxConns: make(chan struct{}, cfg.MaxConns),
	}

	return relay, nil
}

// backendConn represents a pooled connection to a backend
type backendConn struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	mu     sync.Mutex
}

// getBackendConnection establishes or reuses a connection to a backend
func (r *Relay) getBackendConnection(ctx context.Context, backend BackendConfig) (*backendConn, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", backend.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to backend %s: %w", backend.Name, err)
	}

	bc := &backendConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}

	return bc, nil
}

// sendToBackend sends an ICAP request to a backend and returns the response
func (r *Relay) sendToBackend(ctx context.Context, backend BackendConfig, req *ICAPRequest) (*ICAPResponse, error) {
	// Create connection with timeout
	dialCtx, cancel := context.WithTimeout(ctx, time.Duration(r.config.TimeoutMs)*time.Millisecond/2)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", backend.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to backend %s: %w", backend.Name, err)
	}
	defer conn.Close()

	// Set read/write deadlines
	deadline := time.Now().Add(time.Duration(r.config.TimeoutMs) * time.Millisecond)
	conn.SetDeadline(deadline)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Build and send request
	hasBody := len(req.RawBody) > 0
	requestData := BuildICAPRequest(req.Method, backend.Service, req.Headers, req.RawBody, hasBody)
	
	r.logger.Debug("Sending to backend %s: %s", backend.Name, strings.Split(string(requestData), "\r\n")[0])

	if _, err := writer.Write(requestData); err != nil {
		return nil, fmt.Errorf("failed to write to backend %s: %w", backend.Name, err)
	}
	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush to backend %s: %w", backend.Name, err)
	}

	// Parse response
	resp, err := ParseICAPResponse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response from backend %s: %w", backend.Name, err)
	}

	r.logger.Debug("Received from backend %s: ICAP/1.0 %d %s", backend.Name, resp.StatusCode, resp.StatusText)

	return resp, nil
}

// fanOutFanIn implements the fan-out/fan-in pattern with priority logic
func (r *Relay) fanOutFanIn(ctx context.Context, req *ICAPRequest) (*ICAPResponse, error) {
	var primary, secondary BackendConfig
	if r.config.Priority == "A" {
		primary = r.config.BackendA
		secondary = r.config.BackendB
	} else {
		primary = r.config.BackendB
		secondary = r.config.BackendA
	}

	// Channel to receive results
	resultChan := make(chan *BackendResult, 2)

	// Create contexts for cancellation
	primaryCtx, primaryCancel := context.WithTimeout(ctx, time.Duration(r.config.TimeoutMs)*time.Millisecond)
	defer primaryCancel()

	// Start primary request in goroutine
	go func() {
		resp, err := r.sendToBackend(primaryCtx, primary, req)
		resultChan <- &BackendResult{
			Name:     primary.Name,
			Response: resp,
			Error:    err,
		}
	}()

	// Start secondary request in goroutine (only if we haven't gotten a good response yet)
	secondaryCtx, secondaryCancel := context.WithTimeout(ctx, time.Duration(r.config.TimeoutMs)*time.Millisecond)
	defer secondaryCancel()

	go func() {
		// Small delay to give primary a chance to respond first
		select {
		case <-time.After(10 * time.Millisecond):
		case <-secondaryCtx.Done():
			return
		}
		
		resp, err := r.sendToBackend(secondaryCtx, secondary, req)
		resultChan <- &BackendResult{
			Name:     secondary.Name,
			Response: resp,
			Error:    err,
		}
	}()

	// Wait for primary result first
	var primaryResult *BackendResult
	select {
	case result := <-resultChan:
		if result.Name == (func() string {
			if r.config.Priority == "A" {
				return r.config.BackendA.Name
			}
			return r.config.BackendB.Name
		})() {
			primaryResult = result
		} else {
			// Got secondary first, wait for primary
			select {
			case primaryResult = <-resultChan:
			case <-time.After(time.Duration(r.config.TimeoutMs) * time.Millisecond):
				// Primary timed out, use secondary
				r.logger.Info("Primary backend timed out, using secondary")
				return result.Response, result.Error
			}
		}
	case <-time.After(time.Duration(r.config.TimeoutMs) * time.Millisecond):
		return nil, fmt.Errorf("timeout waiting for primary backend")
	}

	// Check if primary succeeded
	if primaryResult.Error == nil && (primaryResult.Response.StatusCode == 200 || primaryResult.Response.StatusCode == 204) {
		r.logger.Info("Using response from primary backend %s (status: %d)", primaryResult.Name, primaryResult.Response.StatusCode)
		// Cancel secondary request
		secondaryCancel()
		return primaryResult.Response, nil
	}

	// Primary failed or returned non-success, wait for secondary
	r.logger.Info("Primary backend %s failed or returned non-success (%v, status: %d), waiting for secondary", 
		primaryResult.Name, primaryResult.Error, 
		func() int {
			if primaryResult.Response != nil {
				return primaryResult.Response.StatusCode
			}
			return 0
		}())

	var secondaryResult *BackendResult
	select {
	case result := <-resultChan:
		secondaryResult = result
	case <-time.After(time.Duration(r.config.TimeoutMs) * time.Millisecond):
		// Secondary also timed out
		r.logger.Error("Both backends timed out")
		return nil, fmt.Errorf("both backends timed out")
	}

	if secondaryResult.Error == nil && (secondaryResult.Response.StatusCode == 200 || secondaryResult.Response.StatusCode == 204) {
		r.logger.Info("Using response from secondary backend %s (status: %d)", secondaryResult.Name, secondaryResult.Response.StatusCode)
		return secondaryResult.Response, nil
	}

	// Both failed
	r.logger.Error("Both backends failed: primary=%v, secondary=%v", primaryResult.Error, secondaryResult.Error)
	if primaryResult.Error != nil {
		return nil, primaryResult.Error
	}
	if secondaryResult.Error != nil {
		return nil, secondaryResult.Error
	}
	
	// Both returned non-success status codes, prefer primary's response
	if primaryResult.Response != nil {
		return primaryResult.Response, nil
	}
	if secondaryResult.Response != nil {
		return secondaryResult.Response, nil
	}
	
	return nil, fmt.Errorf("both backends failed")
}

// handleConnection processes a single client connection
func (r *Relay) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer func() { <-r.maxConns }() // Release semaphore

	r.logger.Info("New connection from %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Create context that cancels on client disconnect
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor connection for close
	go func() {
		buf := make([]byte, 1)
		conn.SetReadDeadline(time.Time{}) // Clear deadline for monitoring
		for {
			_, err := conn.Read(buf)
			if err != nil {
				cancel() // Client disconnected
				return
			}
		}
	}()

	for {
		// Check if context is cancelled (client disconnected)
		select {
		case <-ctx.Done():
			r.logger.Debug("Client disconnected, stopping request processing")
			return
		default:
		}

		// Set read deadline for request
		conn.SetReadDeadline(time.Now().Add(time.Duration(r.config.TimeoutMs) * time.Millisecond))

		// Peek to check if there's data
		_, err := reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				r.logger.Debug("Client closed connection gracefully")
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				r.logger.Debug("Read timeout on client connection")
				return
			}
			r.logger.Error("Error reading from client: %v", err)
			return
		}

		// Parse ICAP request
		req, err := ParseICAPRequest(reader)
		if err != nil {
			r.logger.Error("Failed to parse ICAP request: %v", err)
			// Send error response
			errorResp := BuildICAPResponse(400, "Bad Request", make(http.Header), nil, false)
			writer.Write(errorResp)
			writer.Flush()
			return
		}

		r.logger.Info("Received %s request for service %s", req.Method, req.Service)

		// Handle OPTIONS method specially (healthcheck)
		if req.Method == MethodOPTIONS {
			r.handleOptions(conn, writer, req)
			continue
		}

		// Fan-out to backends with priority logic
		backendResp, err := r.fanOutFanIn(ctx, req)
		
		if err != nil {
			r.logger.Error("Backend processing failed: %v", err)
			// Return 502 Bad Gateway to Squid
			errorHeaders := make(http.Header)
			errorHeaders.Set("Server", "icap-relay")
			errorResp := BuildICAPResponse(502, "Bad Gateway", errorHeaders, nil, false)
			writer.Write(errorResp)
			writer.Flush()
			continue
		}

		// Forward response to client
		responseData := BuildICAPResponse(backendResp.StatusCode, backendResp.StatusText, backendResp.Headers, backendResp.Body, backendResp.IsChunked)
		r.logger.Debug("Sending response to client: ICAP/1.0 %d %s", backendResp.StatusCode, backendResp.StatusText)
		
		if _, err := writer.Write(responseData); err != nil {
			r.logger.Error("Failed to write response to client: %v", err)
			return
		}
		if err := writer.Flush(); err != nil {
			r.logger.Error("Failed to flush response to client: %v", err)
			return
		}
	}
}

// handleOptions handles OPTIONS requests (ICAP healthcheck)
func (r *Relay) handleOptions(conn net.Conn, writer *bufio.Writer, req *ICAPRequest) {
	r.logger.Info("Handling OPTIONS request")

	// For OPTIONS, we can either proxy to backends or respond directly
	// Here we proxy to primary backend to verify it's alive
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.config.TimeoutMs)*time.Millisecond)
	defer cancel()

	var backend BackendConfig
	if r.config.Priority == "A" {
		backend = r.config.BackendA
	} else {
		backend = r.config.BackendB
	}

	resp, err := r.sendToBackend(ctx, backend, req)
	if err != nil {
		r.logger.Error("OPTIONS to backend failed: %v", err)
		// Return a minimal OPTIONS response anyway
		headers := make(http.Header)
		headers.Set("Server", "icap-relay")
		headers.Set("ICAP", "1.0")
		headers.Set("Methods", "REQMOD, RESPMOD, OPTIONS")
		optionsResp := BuildICAPResponse(200, "OK", headers, nil, false)
		writer.Write(optionsResp)
		writer.Flush()
		return
	}

	// Forward backend's OPTIONS response
	responseData := BuildICAPResponse(resp.StatusCode, resp.StatusText, resp.Headers, resp.Body, resp.IsChunked)
	writer.Write(responseData)
	writer.Flush()
	r.logger.Info("OPTIONS response forwarded successfully")
}

// Run starts the relay server
func (r *Relay) Run() error {
	r.logger.Info("Starting ICAP relay on %s", r.config.ListenAddr)
	r.logger.Info("Backend A: %s%s", r.config.BackendA.Address, r.config.BackendA.Service)
	r.logger.Info("Backend B: %s%s", r.config.BackendB.Address, r.config.BackendB.Service)
	r.logger.Info("Priority: %s", r.config.Priority)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		r.logger.Info("Shutdown signal received, closing listener...")
		r.listener.Close()
	}()

	for {
		conn, err := r.listener.Accept()
		if err != nil {
			if ne, ok := err.(*net.OpError); ok && ne.Err.Error() == "use of closed network connection" {
				r.logger.Info("Listener closed, shutting down")
				return nil
			}
			r.logger.Error("Accept error: %v", err)
			continue
		}

		// Check connection limit
		select {
		case r.maxConns <- struct{}{}:
			// Connection accepted
		default:
			r.logger.Info("Max connections reached, rejecting connection from %s", conn.RemoteAddr())
			conn.Close()
			continue
		}

		go r.handleConnection(conn)
	}
}

// Close shuts down the relay
func (r *Relay) Close() error {
	if r.listener != nil {
		return r.listener.Close()
	}
	return nil
}
