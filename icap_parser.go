// === FILE: icap_parser.go ===
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

// ICAPMethod represents ICAP request methods
type ICAPMethod string

const (
	MethodREQMOD  ICAPMethod = "REQMOD"
	MethodRESPMOD ICAPMethod = "RESPMOD"
	MethodOPTIONS ICAPMethod = "OPTIONS"
)

// ICAPRequest represents a parsed ICAP request
type ICAPRequest struct {
	Method     ICAPMethod
	Service    string
	Version    string // e.g., "1.0"
	Headers    http.Header
	Encapsulated map[string]int // section offsets: req-hdr, req-body, res-hdr, res-body, null-body
	RawBody    []byte           // Raw chunked body data
}

// ICAPResponse represents an ICAP response to be sent
type ICAPResponse struct {
	Version    string // e.g., "1.0"
	StatusCode int
	StatusText string
	Headers    http.Header
	Body       []byte
	IsChunked  bool
}

// ParseICAPRequest parses an ICAP request from a reader
func ParseICAPRequest(r *bufio.Reader) (*ICAPRequest, error) {
	req := &ICAPRequest{
		Headers: make(http.Header),
		Encapsulated: make(map[string]int),
	}

	// Read request line: METHOD icap://service ICAP/1.0
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read request line: %w", err)
	}
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid ICAP request line: %s", line)
	}

	req.Method = ICAPMethod(parts[0])
	if req.Method != MethodREQMOD && req.Method != MethodRESPMOD && req.Method != MethodOPTIONS {
		return nil, fmt.Errorf("unsupported ICAP method: %s", req.Method)
	}

	// Parse service URL (icap://host/service)
	serviceURL := parts[1]
	if strings.HasPrefix(serviceURL, "icap://") {
		idx := strings.Index(serviceURL[7:], "/")
		if idx >= 0 {
			req.Service = serviceURL[6+idx:]
		} else {
			req.Service = "/"
		}
	} else {
		req.Service = serviceURL
	}

	req.Version = parts[2]

	// Read headers
	tpReader := textproto.NewReader(r)
	mimeHeader, err := tpReader.ReadMIMEHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}
	req.Headers = http.Header(mimeHeader)

	// Parse Encapsulated header
	if encap := req.Headers.Get("Encapsulated"); encap != "" {
		if err := parseEncapsulated(encap, req.Encapsulated); err != nil {
			return nil, fmt.Errorf("failed to parse Encapsulated header: %w", err)
		}
	}

	// Read body if present (chunked encoding)
	if _, hasBody := req.Encapsulated["req-body"]; hasBody {
		body, err := readChunkedBody(r, -1)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		req.RawBody = body
	} else if _, hasNullBody := req.Encapsulated["null-body"]; hasNullBody {
		// No body expected
		req.RawBody = nil
	}

	return req, nil
}

// parseEncapsulated parses the Encapsulated header value
// Format: "req-hdr=0, req-body=45" or "res-hdr=0, null-body=123"
func parseEncapsulated(value string, result map[string]int) error {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		offset, err := strconv.Atoi(strings.TrimSpace(kv[1]))
		if err != nil {
			return fmt.Errorf("invalid offset in Encapsulated: %s", part)
		}
		result[key] = offset
	}
	return nil
}

// readChunkedBody reads a chunked-encoded body from the reader
func readChunkedBody(r *bufio.Reader, maxSize int) ([]byte, error) {
	var body bytes.Buffer
	totalSize := 0

	for {
		// Read chunk size line
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk size: %w", err)
		}
		line = strings.TrimSpace(line)
		
		// Handle chunk extensions (e.g., "1a;name=value")
		if idx := strings.Index(line, ";"); idx >= 0 {
			line = line[:idx]
		}

		chunkSize, err := strconv.ParseInt(line, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chunk size: %s", line)
		}

		if chunkSize == 0 {
			// End of chunked body, read trailing CRLF and any trailers
			// Read until we get an empty line (end of trailers)
			for {
				trailerLine, err := r.ReadString('\n')
				if err != nil {
					return nil, fmt.Errorf("failed to read trailer: %w", err)
				}
				if strings.TrimSpace(trailerLine) == "" {
					break
				}
			}
			break
		}

		if maxSize > 0 && totalSize+int(chunkSize) > maxSize {
			return nil, fmt.Errorf("body exceeds maximum buffer size")
		}

		// Read chunk data
		chunkData := make([]byte, chunkSize)
		_, err = io.ReadFull(r, chunkData)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk data: %w", err)
		}
		body.Write(chunkData)
		totalSize += int(chunkSize)

		// Read CRLF after chunk
		crlf := make([]byte, 2)
		_, err = io.ReadFull(r, crlf)
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk CRLF: %w", err)
		}
	}

	return body.Bytes(), nil
}

// BuildICAPRequest builds an ICAP request string to send to backend
func BuildICAPRequest(method ICAPMethod, service string, headers http.Header, body []byte, hasBody bool) []byte {
	var buf bytes.Buffer

	// Request line
	buf.WriteString(fmt.Sprintf("%s icap://%s%s ICAP/1.0\r\n", method, headers.Get("Host"), service))

	// Headers
	for key, values := range headers {
		// Skip hop-by-hop headers that should be regenerated
		if key == "Connection" || key == "Keep-Alive" {
			continue
		}
		for _, value := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}

	// Add or update Encapsulated header based on body presence
	if hasBody && len(body) > 0 {
		// Check if Encapsulated already exists
		if encap := headers.Get("Encapsulated"); encap != "" {
			// Keep existing encapsulated structure
			buf.WriteString(fmt.Sprintf("Encapsulated: %s\r\n", encap))
		} else {
			buf.WriteString("Encapsulated: req-hdr=0, req-body=0\r\n")
		}
	} else {
		if headers.Get("Encapsulated") == "" {
			buf.WriteString("Encapsulated: req-hdr=0, null-body=0\r\n")
		}
	}

	buf.WriteString("\r\n")

	// Add body if present (chunked encoding)
	if hasBody && len(body) > 0 {
		// Send body as chunked
		chunkSize := fmt.Sprintf("%x\r\n", len(body))
		buf.WriteString(chunkSize)
		buf.Write(body)
		buf.WriteString("\r\n")
		// Final chunk
		buf.WriteString("0\r\n\r\n")
	}

	return buf.Bytes()
}

// BuildICAPResponse builds an ICAP response to send back to client
func BuildICAPResponse(statusCode int, statusText string, headers http.Header, body []byte, isChunked bool) []byte {
	var buf bytes.Buffer

	// Status line
	buf.WriteString(fmt.Sprintf("ICAP/1.0 %d %s\r\n", statusCode, statusText))

	// Headers
	hasEncapsulated := false
	for key, values := range headers {
		if key == "Encapsulated" {
			hasEncapsulated = true
		}
		if key == "Transfer-Encoding" && !isChunked {
			continue // Don't add TE if not chunked
		}
		for _, value := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}

	// Add Transfer-Encoding if chunked
	if isChunked {
		if headers.Get("Transfer-Encoding") == "" {
			buf.WriteString("Transfer-Encoding: chunked\r\n")
		}
	}

	// Add Encapsulated if not present
	if !hasEncapsulated {
		if len(body) > 0 {
			buf.WriteString("Encapsulated: res-body=0\r\n")
		} else {
			buf.WriteString("Encapsulated: null-body=0\r\n")
		}
	}

	buf.WriteString("\r\n")

	// Add body if present
	if len(body) > 0 {
		if isChunked {
			// Chunked encoding
			chunkSize := fmt.Sprintf("%x\r\n", len(body))
			buf.WriteString(chunkSize)
			buf.Write(body)
			buf.WriteString("\r\n0\r\n\r\n")
		} else {
			buf.Write(body)
		}
	}

	return buf.Bytes()
}

// ParseICAPResponse parses an ICAP response from a reader
func ParseICAPResponse(r *bufio.Reader) (*ICAPResponse, error) {
	resp := &ICAPResponse{
		Headers: make(http.Header),
	}

	// Read status line: ICAP/1.0 200 OK
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read status line: %w", err)
	}
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid ICAP response line: %s", line)
	}

	resp.Version = parts[0]
	resp.StatusCode, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}
	resp.StatusText = strings.Join(parts[2:], " ")

	// Read headers
	tpReader := textproto.NewReader(r)
	mimeHeader, err := tpReader.ReadMIMEHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}
	resp.Headers = http.Header(mimeHeader)

	// Check for chunked encoding
	if resp.Headers.Get("Transfer-Encoding") == "chunked" {
		resp.IsChunked = true
	}

	// Read body if present
	if resp.IsChunked || resp.Headers.Get("Content-Length") != "" {
		body, err := readChunkedBody(r, -1)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		resp.Body = body
	}

	return resp, nil
}
