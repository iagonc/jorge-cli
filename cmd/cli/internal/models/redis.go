package models

// RedisInfo represents Redis server information
type RedisInfo struct {
	Address       string            `json:"address"`
	Connected     bool              `json:"connected"`
	Version       string            `json:"version"`
	Mode          string            `json:"mode"` // standalone, cluster, sentinel
	Role          string            `json:"role"` // master, slave
	Uptime        int64             `json:"uptime_seconds"`
	ConnectedClients int            `json:"connected_clients"`
	UsedMemory    int64             `json:"used_memory"`
	UsedMemoryHuman string          `json:"used_memory_human"`
	MaxMemory     int64             `json:"max_memory"`
	MaxMemoryHuman string           `json:"max_memory_human"`
	MemoryUsagePercent float64      `json:"memory_usage_percent"`
	TotalKeys     int64             `json:"total_keys"`
	ExpiredKeys   int64             `json:"expired_keys"`
	EvictedKeys   int64             `json:"evicted_keys"`
	HitRate       float64           `json:"hit_rate"`
	Databases     map[string]DBInfo `json:"databases"`
	Error         string            `json:"error,omitempty"`
}

// DBInfo represents Redis database info
type DBInfo struct {
	Keys    int64 `json:"keys"`
	Expires int64 `json:"expires"`
	AvgTTL  int64 `json:"avg_ttl"`
}

// RedisKeyInfo represents information about a Redis key
type RedisKeyInfo struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	TTL      int64  `json:"ttl"` // -1 = no expiry, -2 = key doesn't exist
	Size     int64  `json:"size"`
	Encoding string `json:"encoding"`
}

// RedisScanResult represents the result of scanning Redis keys
type RedisScanResult struct {
	Pattern string         `json:"pattern"`
	Keys    []RedisKeyInfo `json:"keys"`
	Total   int            `json:"total"`
}
