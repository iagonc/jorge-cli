package curl

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// CurlUsecase handles HTTP requests with detailed analysis
type CurlUsecase struct {
	logger *zap.Logger
}

// NewCurlUsecase creates a new CurlUsecase
func NewCurlUsecase(logger *zap.Logger) *CurlUsecase {
	return &CurlUsecase{logger: logger}
}

// Execute performs an HTTP request with analysis
func (u *CurlUsecase) Execute(ctx context.Context, config models.CurlRequest) (*models.CurlResult, error) {
	result := &models.CurlResult{
		Redirects: []models.CurlRedirect{},
	}

	if config.Method == "" {
		config.Method = "GET"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 10 * 1024 * 1024 // 10MB
	}

	// Timing variables
	var dnsStart, dnsDone time.Time
	var connectStart, connectDone time.Time
	var tlsStart, tlsDone time.Time
	var firstByte time.Time
	var tlsState *tls.ConnectionState

	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connectDone = time.Now()
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			tlsDone = time.Now()
			tlsState = &state
		},
		GotFirstResponseByte: func() {
			firstByte = time.Now()
		},
	}

	// Create request
	var bodyReader io.Reader
	if config.Body != "" {
		bodyReader = strings.NewReader(config.Body)
	}

	req, err := http.NewRequestWithContext(
		httptrace.WithClientTrace(ctx, trace),
		config.Method,
		config.URL,
		bodyReader,
	)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}

	// Set headers
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "jorge-cli/1.0")
	}

	// Record request info
	result.Request = models.CurlRequestInfo{
		Method:  config.Method,
		URL:     config.URL,
		Headers: make(map[string]string),
		Body:    config.Body,
	}
	for k, v := range req.Header {
		result.Request.Headers[k] = strings.Join(v, ", ")
	}

	// Configure client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Insecure,
		},
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				result.Redirects = append(result.Redirects, models.CurlRedirect{
					From:       via[len(via)-1].URL.String(),
					To:         req.URL.String(),
					StatusCode: 0, // Not available here
				})
			}
			if !config.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	totalTime := time.Since(startTime)

	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	// Read body
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, config.MaxBodySize))
	contentTransferDone := time.Now()

	if err != nil {
		result.Error = fmt.Sprintf("error reading body: %v", err)
		return result, nil
	}

	// Fill timing info
	result.Timing = models.CurlTiming{
		Total: totalTime,
	}
	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		result.Timing.DNSLookup = dnsDone.Sub(dnsStart)
	}
	if !connectStart.IsZero() && !connectDone.IsZero() {
		result.Timing.TCPConnect = connectDone.Sub(connectStart)
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		result.Timing.TLSHandshake = tlsDone.Sub(tlsStart)
	}
	if !firstByte.IsZero() && !connectDone.IsZero() {
		result.Timing.ServerProcessing = firstByte.Sub(connectDone)
		if !tlsDone.IsZero() {
			result.Timing.ServerProcessing = firstByte.Sub(tlsDone)
		}
	}
	if !firstByte.IsZero() {
		result.Timing.ContentTransfer = contentTransferDone.Sub(firstByte)
	}

	// Fill response info
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		respHeaders[k] = strings.Join(v, ", ")
	}

	result.Response = models.CurlResponseInfo{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       respHeaders,
		BodySize:      int64(len(bodyBytes)),
		ContentType:   resp.Header.Get("Content-Type"),
	}

	// Parse cookies
	for _, cookie := range resp.Cookies() {
		result.Response.Cookies = append(result.Response.Cookies, models.CurlCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
		})
	}

	// Include body if requested
	if config.ShowBody {
		result.Response.Body = string(bodyBytes)
	}

	// TLS info
	if tlsState != nil && len(tlsState.PeerCertificates) > 0 {
		cert := tlsState.PeerCertificates[0]
		result.TLS = &models.CurlTLSInfo{
			Version:     tlsVersionName(tlsState.Version),
			CipherSuite: tls.CipherSuiteName(tlsState.CipherSuite),
			Certificate: cert.Subject.CommonName,
			Issuer:      cert.Issuer.CommonName,
		}
	}

	result.Success = true

	return result, nil
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "unknown"
	}
}
