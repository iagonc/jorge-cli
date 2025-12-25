package redis

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// RedisUsecase handles Redis operations
type RedisUsecase struct {
	logger *zap.Logger
}

// NewRedisUsecase creates a new RedisUsecase
func NewRedisUsecase(logger *zap.Logger) *RedisUsecase {
	return &RedisUsecase{logger: logger}
}

// Ping checks if Redis is reachable
func (u *RedisUsecase) Ping(ctx context.Context, address string) (bool, time.Duration, error) {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return false, 0, err
	}
	defer conn.Close()

	// Send PING command
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		return false, 0, err
	}

	// Read response
	buf := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return false, 0, err
	}

	latency := time.Since(start)
	response := string(buf[:n])

	if strings.Contains(response, "PONG") || strings.Contains(response, "+PONG") {
		return true, latency, nil
	}

	return false, latency, fmt.Errorf("unexpected response: %s", response)
}

// GetInfo gets Redis server information
func (u *RedisUsecase) GetInfo(ctx context.Context, address, password string) (*models.RedisInfo, error) {
	info := &models.RedisInfo{
		Address:   address,
		Connected: false,
	}

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		info.Error = err.Error()
		return info, nil
	}
	defer conn.Close()

	// Authenticate if password provided
	if password != "" {
		authCmd := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(password), password)
		conn.Write([]byte(authCmd))

		buf := make([]byte, 128)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, _ := conn.Read(buf)
		if !strings.Contains(string(buf[:n]), "+OK") && !strings.Contains(string(buf[:n]), "OK") {
			info.Error = "Authentication failed"
			return info, nil
		}
	}

	// Send INFO command
	_, err = conn.Write([]byte("*1\r\n$4\r\nINFO\r\n"))
	if err != nil {
		info.Error = err.Error()
		return info, nil
	}

	// Read response
	buf := make([]byte, 16384)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		info.Error = err.Error()
		return info, nil
	}

	info.Connected = true
	response := string(buf[:n])

	// Parse INFO response
	info.Databases = make(map[string]models.DBInfo)

	lines := strings.Split(response, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" || !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "redis_version":
			info.Version = value
		case "redis_mode":
			info.Mode = value
		case "role":
			info.Role = value
		case "uptime_in_seconds":
			info.Uptime, _ = strconv.ParseInt(value, 10, 64)
		case "connected_clients":
			info.ConnectedClients, _ = strconv.Atoi(value)
		case "used_memory":
			info.UsedMemory, _ = strconv.ParseInt(value, 10, 64)
		case "used_memory_human":
			info.UsedMemoryHuman = value
		case "maxmemory":
			info.MaxMemory, _ = strconv.ParseInt(value, 10, 64)
		case "maxmemory_human":
			info.MaxMemoryHuman = value
		case "keyspace_hits":
			hits, _ := strconv.ParseFloat(value, 64)
			// Will calculate hit rate after getting misses
			info.HitRate = hits
		case "keyspace_misses":
			misses, _ := strconv.ParseFloat(value, 64)
			if info.HitRate+misses > 0 {
				info.HitRate = (info.HitRate / (info.HitRate + misses)) * 100
			}
		case "expired_keys":
			info.ExpiredKeys, _ = strconv.ParseInt(value, 10, 64)
		case "evicted_keys":
			info.EvictedKeys, _ = strconv.ParseInt(value, 10, 64)
		}

		// Parse database info (db0, db1, etc.)
		if strings.HasPrefix(key, "db") {
			dbInfo := models.DBInfo{}
			for _, kv := range strings.Split(value, ",") {
				kvParts := strings.Split(kv, "=")
				if len(kvParts) == 2 {
					switch kvParts[0] {
					case "keys":
						dbInfo.Keys, _ = strconv.ParseInt(kvParts[1], 10, 64)
						info.TotalKeys += dbInfo.Keys
					case "expires":
						dbInfo.Expires, _ = strconv.ParseInt(kvParts[1], 10, 64)
					case "avg_ttl":
						dbInfo.AvgTTL, _ = strconv.ParseInt(kvParts[1], 10, 64)
					}
				}
			}
			info.Databases[key] = dbInfo
		}
	}

	// Calculate memory usage percentage
	if info.MaxMemory > 0 {
		info.MemoryUsagePercent = float64(info.UsedMemory) / float64(info.MaxMemory) * 100
	}

	return info, nil
}

// ScanKeys scans Redis keys matching a pattern
func (u *RedisUsecase) ScanKeys(ctx context.Context, address, password, pattern string, count int) (*models.RedisScanResult, error) {
	result := &models.RedisScanResult{
		Pattern: pattern,
		Keys:    []models.RedisKeyInfo{},
	}

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Authenticate if password provided
	if password != "" {
		authCmd := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(password), password)
		conn.Write([]byte(authCmd))

		buf := make([]byte, 128)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, _ := conn.Read(buf)
		if !strings.Contains(string(buf[:n]), "+OK") {
			return nil, fmt.Errorf("authentication failed")
		}
	}

	// Use KEYS command for simplicity (SCAN would be better for production)
	keysCmd := fmt.Sprintf("*2\r\n$4\r\nKEYS\r\n$%d\r\n%s\r\n", len(pattern), pattern)
	_, err = conn.Write([]byte(keysCmd))
	if err != nil {
		return nil, err
	}

	// Read response (simplified parsing)
	buf := make([]byte, 65536)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	// Parse array response (simplified)
	response := string(buf[:n])
	lines := strings.Split(response, "\r\n")

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "$") {
			continue
		}

		if len(result.Keys) >= count {
			break
		}

		result.Keys = append(result.Keys, models.RedisKeyInfo{
			Key: line,
		})
	}

	result.Total = len(result.Keys)
	return result, nil
}

// FlushDB flushes a Redis database
func (u *RedisUsecase) FlushDB(ctx context.Context, address, password string, db int) error {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Authenticate if password provided
	if password != "" {
		authCmd := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(password), password)
		conn.Write([]byte(authCmd))

		buf := make([]byte, 128)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, _ := conn.Read(buf)
		if !strings.Contains(string(buf[:n]), "+OK") {
			return fmt.Errorf("authentication failed")
		}
	}

	// Select database
	selectCmd := fmt.Sprintf("*2\r\n$6\r\nSELECT\r\n$%d\r\n%d\r\n", len(strconv.Itoa(db)), db)
	conn.Write([]byte(selectCmd))

	buf := make([]byte, 128)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.Read(buf)

	// Flush
	_, err = conn.Write([]byte("*1\r\n$7\r\nFLUSHDB\r\n"))
	if err != nil {
		return err
	}

	buf = make([]byte, 128)
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if strings.Contains(string(buf[:n]), "+OK") {
		return nil
	}

	return fmt.Errorf("flush failed: %s", string(buf[:n]))
}
