package playbook

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// BuiltinFunc defines the signature for built-in functions
type BuiltinFunc func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// BuiltinRegistry manages available built-in functions
type BuiltinRegistry struct {
	functions map[string]BuiltinFunc
	logger    *zap.Logger
}

// NewBuiltinRegistry creates a new BuiltinRegistry
func NewBuiltinRegistry(logger *zap.Logger) *BuiltinRegistry {
	r := &BuiltinRegistry{
		functions: make(map[string]BuiltinFunc),
		logger:    logger,
	}
	r.registerDefaults()
	return r
}

func (r *BuiltinRegistry) registerDefaults() {
	r.functions["dns_lookup"] = r.dnsLookup
	r.functions["ping"] = r.ping
	r.functions["http_request"] = r.httpRequest
	r.functions["port_check"] = r.portCheck
	r.functions["ssl_check"] = r.sslCheck
	r.functions["traceroute"] = r.traceroute
	r.functions["curl"] = r.curl
	r.functions["sleep"] = r.sleep
	r.functions["echo"] = r.echo
}

// Get returns a builtin function by name
func (r *BuiltinRegistry) Get(name string) (BuiltinFunc, bool) {
	fn, ok := r.functions[name]
	return fn, ok
}

// List returns all available function names
func (r *BuiltinRegistry) List() []string {
	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}

func (r *BuiltinRegistry) dnsLookup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	domain, ok := params["domain"].(string)
	if !ok || domain == "" {
		return nil, fmt.Errorf("domain parameter is required")
	}

	resolver := net.DefaultResolver
	ips, err := resolver.LookupHost(ctx, domain)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"domain": domain,
		"ips":    ips,
	}
	if len(ips) > 0 {
		result["ip"] = ips[0]
	}

	return result, nil
}

func (r *BuiltinRegistry) ping(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	host, ok := params["host"].(string)
	if !ok || host == "" {
		return nil, fmt.Errorf("host parameter is required")
	}

	count := 4
	if c, ok := params["count"].(int); ok {
		count = c
	}

	cmd := exec.CommandContext(ctx, "ping", "-c", fmt.Sprintf("%d", count), host)
	output, err := cmd.Output()

	result := map[string]interface{}{
		"host":    host,
		"success": err == nil,
		"output":  string(output),
	}

	return result, nil
}

func (r *BuiltinRegistry) httpRequest(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	url, ok := params["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	method := "GET"
	if m, ok := params["method"].(string); ok {
		method = m
	}

	timeout := 10 * time.Second
	if t, ok := params["timeout"].(int); ok {
		timeout = time.Duration(t) * time.Second
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	result := map[string]interface{}{
		"url":           url,
		"method":        method,
		"response_time": duration.Milliseconds(),
	}

	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	result["success"] = true
	result["status_code"] = resp.StatusCode
	result["status"] = resp.Status

	// Check expected status
	if expectedStatus, ok := params["expected_status"].(int); ok {
		result["match"] = resp.StatusCode == expectedStatus
	}

	return result, nil
}

func (r *BuiltinRegistry) portCheck(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	host, ok := params["host"].(string)
	if !ok || host == "" {
		return nil, fmt.Errorf("host parameter is required")
	}

	port, ok := params["port"].(int)
	if !ok {
		return nil, fmt.Errorf("port parameter is required")
	}

	timeout := 5 * time.Second
	if t, ok := params["timeout"].(int); ok {
		timeout = time.Duration(t) * time.Second
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)

	result := map[string]interface{}{
		"host": host,
		"port": port,
		"open": err == nil,
	}

	if err == nil {
		conn.Close()
	} else {
		result["error"] = err.Error()
	}

	return result, nil
}

func (r *BuiltinRegistry) sslCheck(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	domain, ok := params["domain"].(string)
	if !ok || domain == "" {
		return nil, fmt.Errorf("domain parameter is required")
	}

	port := 443
	if p, ok := params["port"].(int); ok {
		port = p
	}

	// Use openssl for SSL check
	cmd := exec.CommandContext(ctx, "openssl", "s_client", "-connect",
		fmt.Sprintf("%s:%d", domain, port), "-servername", domain)
	cmd.Stdin = strings.NewReader("")

	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"domain": domain,
		"port":   port,
		"valid":  err == nil && strings.Contains(string(output), "Verify return code: 0"),
		"output": string(output),
	}

	// Try to extract expiry info
	if strings.Contains(string(output), "notAfter=") {
		result["has_expiry_info"] = true
	}

	return result, nil
}

func (r *BuiltinRegistry) traceroute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	host, ok := params["host"].(string)
	if !ok || host == "" {
		return nil, fmt.Errorf("host parameter is required")
	}

	maxHops := 15
	if m, ok := params["max_hops"].(int); ok {
		maxHops = m
	}

	cmd := exec.CommandContext(ctx, "traceroute", "-m", fmt.Sprintf("%d", maxHops), host)
	output, err := cmd.Output()

	result := map[string]interface{}{
		"host":    host,
		"success": err == nil,
		"output":  string(output),
	}

	return result, nil
}

func (r *BuiltinRegistry) curl(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	url, ok := params["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	args := []string{"-s", "-o", "/dev/null", "-w", "%{http_code} %{time_total}", url}

	if method, ok := params["method"].(string); ok {
		args = append([]string{"-X", method}, args...)
	}

	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.Output()

	result := map[string]interface{}{
		"url":     url,
		"success": err == nil,
		"output":  string(output),
	}

	// Parse output
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		result["status_code"] = parts[0]
		result["response_time"] = parts[1]
	}

	return result, nil
}

func (r *BuiltinRegistry) sleep(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	duration := 1 * time.Second

	if d, ok := params["seconds"].(int); ok {
		duration = time.Duration(d) * time.Second
	} else if d, ok := params["duration"].(string); ok {
		parsed, err := time.ParseDuration(d)
		if err == nil {
			duration = parsed
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(duration):
		return map[string]interface{}{"slept": duration.String()}, nil
	}
}

func (r *BuiltinRegistry) echo(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	message, _ := params["message"].(string)
	return map[string]interface{}{"message": message}, nil
}
