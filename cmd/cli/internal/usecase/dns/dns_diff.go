package dns

import (
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// DNSResolver interface for testability
type DNSResolver interface {
	LookupHost(ctx context.Context, host string) ([]string, error)
	LookupCNAME(ctx context.Context, host string) (string, error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
}

// DefaultDNSResolver implements DNSResolver using net.Resolver
type DefaultDNSResolver struct {
	Resolver *net.Resolver
}

func (r *DefaultDNSResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	return r.Resolver.LookupHost(ctx, host)
}

func (r *DefaultDNSResolver) LookupCNAME(ctx context.Context, host string) (string, error) {
	return r.Resolver.LookupCNAME(ctx, host)
}

func (r *DefaultDNSResolver) LookupMX(ctx context.Context, name string) ([]*net.MX, error) {
	return r.Resolver.LookupMX(ctx, name)
}

func (r *DefaultDNSResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	return r.Resolver.LookupTXT(ctx, name)
}

func (r *DefaultDNSResolver) LookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	return r.Resolver.LookupNS(ctx, name)
}

// DNSDiffUsecase handles DNS comparison operations
type DNSDiffUsecase struct {
	Logger   *zap.Logger
	Resolver DNSResolver
}

// NewDNSDiffUsecase creates a new DNSDiffUsecase
func NewDNSDiffUsecase(logger *zap.Logger) *DNSDiffUsecase {
	return &DNSDiffUsecase{
		Logger: logger,
		Resolver: &DefaultDNSResolver{
			Resolver: net.DefaultResolver,
		},
	}
}

// NewDNSDiffUsecaseWithNameserver creates a DNSDiffUsecase with a custom nameserver
func NewDNSDiffUsecaseWithNameserver(logger *zap.Logger, nameserver string) *DNSDiffUsecase {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, "udp", nameserver+":53")
		},
	}

	return &DNSDiffUsecase{
		Logger: logger,
		Resolver: &DefaultDNSResolver{
			Resolver: resolver,
		},
	}
}

// LoadConfig loads DNS diff configuration from a YAML file
func (u *DNSDiffUsecase) LoadConfig(configPath string) (*models.DNSDiffConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config models.DNSDiffConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// CompareDNS compares expected DNS records with actual records
func (u *DNSDiffUsecase) CompareDNS(ctx context.Context, config *models.DNSDiffConfig) (*models.DNSDiffResult, []error) {
	u.Logger.Info("Starting DNS comparison", zap.Int("records", len(config.Records)))

	startTime := time.Now()
	result := &models.DNSDiffResult{
		Nameserver:   config.Nameserver,
		TotalRecords: len(config.Records),
	}

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		errorsList []error
	)

	for _, record := range config.Records {
		wg.Add(1)
		go func(r models.ExpectedRecord) {
			defer wg.Done()

			diff := u.compareRecord(ctx, r)

			mu.Lock()
			result.Diffs = append(result.Diffs, diff)
			if diff.Error != "" {
				result.Errors++
				errorsList = append(errorsList, fmt.Errorf("%s %s: %s", r.Domain, r.RecordType, diff.Error))
			} else if diff.Match {
				result.Matching++
			} else {
				result.Mismatched++
			}
			mu.Unlock()
		}(record)
	}

	wg.Wait()

	result.Duration = time.Since(startTime)

	u.Logger.Info("DNS comparison completed",
		zap.Int("matching", result.Matching),
		zap.Int("mismatched", result.Mismatched),
		zap.Int("errors", result.Errors))

	return result, errorsList
}

func (u *DNSDiffUsecase) compareRecord(ctx context.Context, record models.ExpectedRecord) models.RecordDiff {
	diff := models.RecordDiff{
		Domain:     record.Domain,
		RecordType: record.RecordType,
		Expected:   record.Values,
	}

	actual, err := u.lookupRecord(ctx, record.Domain, record.RecordType)
	if err != nil {
		diff.Error = err.Error()
		return diff
	}

	diff.Actual = actual
	diff.Match, diff.Missing, diff.Extra = compareSlices(record.Values, actual)

	return diff
}

func (u *DNSDiffUsecase) lookupRecord(ctx context.Context, domain string, recordType models.DNSRecordType) ([]string, error) {
	switch recordType {
	case models.RecordTypeA, models.RecordTypeAAAA:
		hosts, err := u.Resolver.LookupHost(ctx, domain)
		if err != nil {
			return nil, err
		}
		// Filter based on record type
		var filtered []string
		for _, h := range hosts {
			ip := net.ParseIP(h)
			if ip == nil {
				continue
			}
			if recordType == models.RecordTypeA && ip.To4() != nil {
				filtered = append(filtered, h)
			} else if recordType == models.RecordTypeAAAA && ip.To4() == nil {
				filtered = append(filtered, h)
			}
		}
		return filtered, nil

	case models.RecordTypeCNAME:
		cname, err := u.Resolver.LookupCNAME(ctx, domain)
		if err != nil {
			return nil, err
		}
		return []string{strings.TrimSuffix(cname, ".")}, nil

	case models.RecordTypeMX:
		mxs, err := u.Resolver.LookupMX(ctx, domain)
		if err != nil {
			return nil, err
		}
		var result []string
		for _, mx := range mxs {
			result = append(result, fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(mx.Host, ".")))
		}
		return result, nil

	case models.RecordTypeTXT:
		return u.Resolver.LookupTXT(ctx, domain)

	case models.RecordTypeNS:
		nss, err := u.Resolver.LookupNS(ctx, domain)
		if err != nil {
			return nil, err
		}
		var result []string
		for _, ns := range nss {
			result = append(result, strings.TrimSuffix(ns.Host, "."))
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported record type: %s", recordType)
	}
}

// CheckSingleRecord checks a single DNS record against expected value
func (u *DNSDiffUsecase) CheckSingleRecord(ctx context.Context, domain string, recordType models.DNSRecordType, expected string) (*models.RecordDiff, error) {
	record := models.ExpectedRecord{
		Domain:     domain,
		RecordType: recordType,
		Values:     []string{expected},
	}

	diff := u.compareRecord(ctx, record)
	if diff.Error != "" {
		return &diff, fmt.Errorf(diff.Error)
	}

	return &diff, nil
}

func compareSlices(expected, actual []string) (match bool, missing, extra []string) {
	expectedSet := make(map[string]bool)
	actualSet := make(map[string]bool)

	for _, e := range expected {
		expectedSet[strings.ToLower(e)] = true
	}
	for _, a := range actual {
		actualSet[strings.ToLower(a)] = true
	}

	// Find missing (in expected but not in actual)
	for e := range expectedSet {
		if !actualSet[e] {
			missing = append(missing, e)
		}
	}

	// Find extra (in actual but not in expected)
	for a := range actualSet {
		if !expectedSet[a] {
			extra = append(extra, a)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)

	match = len(missing) == 0 && len(extra) == 0
	return
}
