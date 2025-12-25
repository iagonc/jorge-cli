package trace

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// TraceUsecase handles HTTP request tracing
type TraceUsecase struct {
	logger *zap.Logger
}

// NewTraceUsecase creates a new TraceUsecase
func NewTraceUsecase(logger *zap.Logger) *TraceUsecase {
	return &TraceUsecase{logger: logger}
}

// Trace performs a traced HTTP request
func (u *TraceUsecase) Trace(ctx context.Context, targetURL, method string, headers map[string]string, body string, showBody bool) (*models.TraceResult, error) {
	if method == "" {
		method = "GET"
	}

	if _, err := url.Parse(targetURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	result := &models.TraceResult{
		URL:       targetURL,
		Method:    method,
		Phases:    []models.TracePhase{},
		Redirects: []models.TraceRedirect{},
	}

	var dnsStart, dnsDone time.Time
	var connectStart, connectDone time.Time
	var tlsStart, tlsDone time.Time
	var gotConn time.Time
	var firstByte time.Time
	var tlsState *tls.ConnectionState

	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
			phase := models.TracePhase{
				Name:      "DNS Lookup",
				StartTime: dnsStart,
				EndTime:   dnsDone,
				Duration:  dnsDone.Sub(dnsStart),
				Success:   info.Err == nil,
			}
			if info.Err != nil {
				phase.Error = info.Err.Error()
			} else if len(info.Addrs) > 0 {
				var ips []string
				for _, addr := range info.Addrs {
					ips = append(ips, addr.IP.String())
				}
				phase.Details = strings.Join(ips, ", ")
			}
			result.Phases = append(result.Phases, phase)
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connectDone = time.Now()
			phase := models.TracePhase{
				Name:      "TCP Connect",
				StartTime: connectStart,
				EndTime:   connectDone,
				Duration:  connectDone.Sub(connectStart),
				Success:   err == nil,
				Details:   addr,
			}
			if err != nil {
				phase.Error = err.Error()
			}
			result.Phases = append(result.Phases, phase)
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			tlsDone = time.Now()
			tlsState = &state
			phase := models.TracePhase{
				Name:      "TLS Handshake",
				StartTime: tlsStart,
				EndTime:   tlsDone,
				Duration:  tlsDone.Sub(tlsStart),
				Success:   err == nil,
				Details:   fmt.Sprintf("TLS %s", tlsVersionName(state.Version)),
			}
			if err != nil {
				phase.Error = err.Error()
			}
			result.Phases = append(result.Phases, phase)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			gotConn = time.Now()
		},
		GotFirstResponseByte: func() {
			firstByte = time.Now()
			phase := models.TracePhase{
				Name:      "Server Processing",
				StartTime: gotConn,
				EndTime:   firstByte,
				Duration:  firstByte.Sub(gotConn),
				Success:   true,
				Details:   "Time to first byte",
			}
			result.Phases = append(result.Phases, phase)
		},
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), method, targetURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "jorge-cli/1.0")
	}

	// Track redirects
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				result.Redirects = append(result.Redirects, models.TraceRedirect{
					From:       via[len(via)-1].URL.String(),
					To:         req.URL.String(),
					StatusCode: 0, // Will be filled later
				})
			}
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	totalTime := time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
		result.TotalDuration = totalTime
		return result, nil
	}
	defer resp.Body.Close()

	// Read body
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 1MB limit
	contentTransferDone := time.Now()

	// Content transfer phase
	if !firstByte.IsZero() {
		phase := models.TracePhase{
			Name:      "Content Transfer",
			StartTime: firstByte,
			EndTime:   contentTransferDone,
			Duration:  contentTransferDone.Sub(firstByte),
			Success:   true,
			Details:   fmt.Sprintf("%d bytes", len(bodyBytes)),
		}
		result.Phases = append(result.Phases, phase)
	}

	result.Success = true
	result.StatusCode = resp.StatusCode
	result.TotalDuration = totalTime

	// Response info
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		respHeaders[k] = strings.Join(v, ", ")
	}

	result.Response = &models.TraceResponse{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       respHeaders,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
	}

	if showBody && len(bodyBytes) > 0 {
		result.Response.Body = string(bodyBytes)
	}

	// TLS info
	if tlsState != nil && len(tlsState.PeerCertificates) > 0 {
		cert := tlsState.PeerCertificates[0]
		result.TLS = &models.TraceTLSInfo{
			Version:            tlsVersionName(tlsState.Version),
			CipherSuite:        tls.CipherSuiteName(tlsState.CipherSuite),
			ServerName:         tlsState.ServerName,
			CertificateSubject: cert.Subject.CommonName,
			CertificateIssuer:  cert.Issuer.CommonName,
			CertificateExpiry:  cert.NotAfter,
			HandshakeDuration:  tlsDone.Sub(tlsStart),
		}
	}

	// Set redirect status codes
	for i := range result.Redirects {
		if i == 0 {
			result.Redirects[i].StatusCode = 301 // Assume 301, not tracked precisely
		}
	}

	return result, nil
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return "unknown"
	}
}
