package network

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

type NetworkDebugUsecase struct {
    Logger *zap.Logger
}

func NewNetworkDebugUsecase(logger *zap.Logger) *NetworkDebugUsecase {
    return &NetworkDebugUsecase{
        Logger: logger,
    }
}

func (u *NetworkDebugUsecase) NetworkDebug(ctx context.Context, domain string) (*models.NetworkDebugResult, []error) {
    result := &models.NetworkDebugResult{}
    var wg sync.WaitGroup
    var mu sync.Mutex
    var errorsList []error

    // Define a helper function to execute a tool and handle results/errors
    executeTool := func(toolName string, fn func() error) {
        defer wg.Done()
        if err := fn(); err != nil {
            mu.Lock()
            errorsList = append(errorsList, fmt.Errorf("%s error: %w", toolName, err))
            mu.Unlock()
        }
    }

    wg.Add(7)

    // Execute tools concurrently
    go executeTool("dig", func() error {
        dns, err := runDig(domain)
        if err != nil {
            return err
        }
        mu.Lock()
        result.DNSLookup = dns
        mu.Unlock()
        return nil
    })

    go executeTool("nslookup", func() error {
        ns, err := runNSLookup(domain)
        if err != nil {
            return err
        }
        mu.Lock()
        result.NSLookup = ns
        mu.Unlock()
        return nil
    })

    go executeTool("traceroute", func() error {
        tr, err := runTraceroute(domain)
        if err != nil {
            return err
        }
        mu.Lock()
        result.Traceroute = tr
        mu.Unlock()
        return nil
    })

    go executeTool("curl", func() error {
        curl, err := runCurl(domain)
        if err != nil {
            return err
        }
        mu.Lock()
        result.HTTPRequest = curl
        mu.Unlock()
        return nil
    })

    go executeTool("ping", func() error {
        ping, err := runPing(domain)
        if err != nil {
            return err
        }
        mu.Lock()
        result.Ping = ping
        mu.Unlock()
        return nil
    })

    go executeTool("netstat", func() error {
        netstat, err := runNetstat()
        if err != nil {
            return err
        }
        mu.Lock()
        result.Netstat = netstat
        mu.Unlock()
        return nil
    })

    go executeTool("iftop", func() error {
        // TODO: Assuming eth0; consider making this configurable
        iftop, err := runIftop("eth0")
        if err != nil {
            return err
        }
        mu.Lock()
        result.Iftop = iftop
        mu.Unlock()
        return nil
    })

    wg.Wait()

    return result, errorsList
}

// Helper functions to execute network tools

func runDig(domain string) (models.DNSLookupResult, error) {
    cmd := exec.Command("dig", "+noall", "+answer", domain)
    output, err := cmd.Output()
    if err != nil {
        return models.DNSLookupResult{}, err
    }

    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    var records []models.DNSRecord
    for _, line := range lines {
        parts := strings.Fields(line)
        if len(parts) < 5 {
            continue
        }
        // parts[3] is the record type, parts[4] is the IP
        recordType := parts[3]
        ip := parts[4]
        records = append(records, models.DNSRecord{
            Type: recordType,
            IP:   ip,
        })
    }

    return models.DNSLookupResult{
        Records: records,
    }, nil
}

func runNSLookup(domain string) (models.NSLookupResult, error) {
    cmd := exec.Command("nslookup", domain)
    output, err := cmd.Output()
    if err != nil {
        return models.NSLookupResult{}, err
    }

    lines := strings.Split(string(output), "\n")
    var ip string
    for _, line := range lines {
        if strings.Contains(line, "Address:") && !strings.Contains(line, "#") {
            parts := strings.Split(line, ":")
            if len(parts) > 1 {
                ip = strings.TrimSpace(parts[1])
                break
            }
        }
    }

    if ip == "" {
        return models.NSLookupResult{}, fmt.Errorf("no IP address found")
    }

    return models.NSLookupResult{
        IP: ip,
    }, nil
}

func runTraceroute(domain string) (models.TracerouteResult, error) {
    cmd := exec.Command("traceroute", "-m", "5", domain)
    output, err := cmd.Output()
    if err != nil {
        return models.TracerouteResult{}, err
    }

    lines := strings.Split(string(output), "\n")
    var hops []models.TracerouteHop
    for _, line := range lines[1:] { // Skip the first line
        if line == "" {
            continue
        }
        parts := strings.Fields(line)
        if len(parts) < 3 {
            continue
        }
        hopNumber, err := strconv.Atoi(parts[0])
        if err != nil {
            continue
        }
        address := parts[1]
        responseTime := parts[len(parts)-2] // Assuming the time is before the last "ms"

        hops = append(hops, models.TracerouteHop{
            HopNumber:    hopNumber,
            Address:      address,
            ResponseTime: responseTime + " ms",
        })
    }

    return models.TracerouteResult{
        Hops: hops,
    }, nil
}

func runCurl(domain string) (models.HTTPRequestResult, error) {
    start := time.Now()
    cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code} %{time_total} %{content_type}", domain)
    output, err := cmd.Output()
    if err != nil {
        return models.HTTPRequestResult{}, err
    }

    parts := strings.Fields(string(output))
    if len(parts) < 3 {
        return models.HTTPRequestResult{}, fmt.Errorf("unexpected curl output: %s", string(output))
    }

    status := parts[0]
    responseTime := fmt.Sprintf("%.0f ms", time.Since(start).Seconds()*1000)
    contentType := parts[2]

    return models.HTTPRequestResult{
        Status:       fmt.Sprintf("HTTP %s", status),
        ResponseTime: responseTime,
        ContentType:  contentType,
    }, nil
}

func runPing(domain string) (models.PingResult, error) {
    cmd := exec.Command("ping", "-c", "4", domain)
    output, err := cmd.Output()
    if err != nil {
        return models.PingResult{}, err
    }

    var sent, received int
    var lossPercent float64
    var avgLatency int

    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        if strings.Contains(line, "packets transmitted") {
            // Example: 4 packets transmitted, 4 received, 0% packet loss, time 3005ms
            parts := strings.Split(line, ",")
            if len(parts) >= 3 {
                fmt.Sscanf(parts[0], "%d packets transmitted", &sent)
                fmt.Sscanf(parts[1], " %d received", &received)
                fmt.Sscanf(parts[2], " %f%% packet loss", &lossPercent)
            }
        }
        if strings.Contains(line, "rtt min/avg/max/mdev") {
            // Example: rtt min/avg/max/mdev = 10.123/15.456/20.789/2.345 ms
            parts := strings.Split(line, "=")
            if len(parts) == 2 {
                stats := strings.Split(strings.TrimSpace(parts[1]), "/")
                if len(stats) >= 2 {
                    avg, err := strconv.ParseFloat(stats[1], 64)
                    if err == nil {
                        avgLatency = int(avg)
                    }
                }
            }
        }
    }

    return models.PingResult{
        Sent:         sent,
        Received:     received,
        Lost:         sent - received,
        LossPercent:  lossPercent,
        AvgLatency:   avgLatency,
    }, nil
}

func runNetstat() (models.NetstatResult, error) {
    cmd := exec.Command("netstat", "-tunapl")
    output, err := cmd.Output()
    if err != nil {
        return models.NetstatResult{}, err
    }

    lines := strings.Split(string(output), "\n")
    var connections []models.NetstatConnection
    for _, line := range lines {
        if strings.HasPrefix(line, "tcp") || strings.HasPrefix(line, "udp") {
            parts := strings.Fields(line)
            if len(parts) >= 6 {
                connections = append(connections, models.NetstatConnection{
                    Protocol:      parts[0],
                    LocalAddress:  parts[3],
                    RemoteAddress: parts[4],
                    Status:        parts[5],
                })
            }
        }
    }

    return models.NetstatResult{
        Connections: connections,
    }, nil
}

func runIftop(interfaceName string) (models.IftopResult, error) {
    // Note: iftop typically requires root privileges. Ensure the CLI has necessary permissions.
    // We'll run iftop in text mode for 5 seconds and capture the output.
    cmd := exec.Command("sudo", "iftop", "-t", "-s", "5", "-i", interfaceName)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return models.IftopResult{}, err
    }

    lines := strings.Split(string(out.String()), "\n")
    var sending, receiving string
    var topConns []models.IftopConnection

    for _, line := range lines {
        if strings.Contains(line, "=>") || strings.Contains(line, "<=") {
            parts := strings.Fields(line)
            if len(parts) >= 6 {
                topConns = append(topConns, models.IftopConnection{
                    Source:        parts[0],
                    Destination:   parts[2],
                    SentKBps:      parts[4],
                    ReceivedKBps: parts[5],
                })
            }
        }
        if strings.Contains(line, "Total send rate") {
            // Example: Total send rate: 120.00 KB/s
            parts := strings.Split(line, ":")
            if len(parts) == 2 {
                sending = strings.TrimSpace(strings.TrimSuffix(parts[1], " KB/s"))
            }
        }
        if strings.Contains(line, "Total receive rate") {
            // Example: Total receive rate: 250.00 KB/s
            parts := strings.Split(line, ":")
            if len(parts) == 2 {
                receiving = strings.TrimSpace(strings.TrimSuffix(parts[1], " KB/s"))
            }
        }
    }

    // Select Top 3 connections
    limit := 3
    if len(topConns) < 3 {
        limit = len(topConns)
    }
    topConns = topConns[:limit]

    return models.IftopResult{
        SendingKBps:    sending + " KB/s",
        ReceivingKBps:  receiving + " KB/s",
        TopConnections: topConns,
    }, nil
}
