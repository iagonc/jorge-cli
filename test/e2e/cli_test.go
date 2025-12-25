package e2e_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/iagonc/jorge-cli/test/testutil"
)

// TestSSLCheck_RealEndpoint tests SSL check against a real endpoint
func TestSSLCheck_RealEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	// Test directly using crypto/tls since we can't import internal packages
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dialer := &tls.Dialer{
		Config: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	conn, err := dialer.DialContext(ctx, "tcp", "google.com:443")
	if err != nil {
		t.Fatalf("SSL connection failed: %v", err)
	}
	defer conn.Close()

	tlsConn := conn.(*tls.Conn)
	state := tlsConn.ConnectionState()

	if len(state.PeerCertificates) == 0 {
		t.Fatal("No certificates received")
	}

	cert := state.PeerCertificates[0]
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

	if daysUntilExpiry <= 0 {
		t.Error("Expected positive days until expiry")
	}

	t.Logf("Certificate expires in %d days", daysUntilExpiry)
}

// TestHealthCheck_WithTestServer tests health check against a test server
func TestHealthCheck_WithTestServer(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test. Set RUN_E2E_TESTS=true to run.")
	}

	ts := testutil.NewTestServer(t)
	ts.SeedData()

	// Test using standard HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(ts.BaseURL + "/resources")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Logf("Health check successful, status: %d", resp.StatusCode)
}

// TestCLI_Help tests that the CLI help command works
func TestCLI_Help(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test. Set RUN_E2E_TESTS=true to run.")
	}

	// Build the CLI first
	buildCmd := exec.Command("go", "build", "-o", "/tmp/jorge-cli-test", "./cmd/cli")
	buildCmd.Dir = getProjectRoot()
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer os.Remove("/tmp/jorge-cli-test")

	// Run help command
	cmd := exec.Command("/tmp/jorge-cli-test", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI help failed: %v, output: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "ssl-check") {
		t.Error("Expected help to mention ssl-check command")
	}
	if !strings.Contains(outputStr, "port-scan") {
		t.Error("Expected help to mention port-scan command")
	}
	if !strings.Contains(outputStr, "health-check") {
		t.Error("Expected help to mention health-check command")
	}
}

// TestCLI_SSLCheck tests the ssl-check command against a real endpoint
func TestCLI_SSLCheck(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E test. Set RUN_E2E_TESTS=true to run.")
	}

	// Build the CLI first
	buildCmd := exec.Command("go", "build", "-o", "/tmp/jorge-cli-test", "./cmd/cli")
	buildCmd.Dir = getProjectRoot()
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer os.Remove("/tmp/jorge-cli-test")

	// Run ssl-check command
	cmd := exec.Command("/tmp/jorge-cli-test", "ssl-check", "--host", "google.com")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("SSL check failed: %v, output: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Certificate") {
		t.Errorf("Expected output to contain certificate info, got: %s", outputStr)
	}
}

func getProjectRoot() string {
	// This assumes tests are run from the project root
	wd, _ := os.Getwd()
	// Navigate up if we're in a subdirectory
	if strings.HasSuffix(wd, "/test/e2e") {
		return wd[:len(wd)-len("/test/e2e")]
	}
	return wd
}
