package models

import "time"

// MTRResult represents the result of an MTR (My Traceroute) analysis
type MTRResult struct {
	Target      string        `json:"target"`
	TargetIP    string        `json:"target_ip"`
	Hops        []MTRHop      `json:"hops"`
	TotalHops   int           `json:"total_hops"`
	PacketsSent int           `json:"packets_sent"`
	Duration    time.Duration `json:"duration"`
	Completed   bool          `json:"completed"`
	Summary     MTRSummary    `json:"summary"`
}

// MTRHop represents a single hop in the network path
type MTRHop struct {
	Number      int           `json:"number"`
	Host        string        `json:"host"`
	IP          string        `json:"ip"`
	Loss        float64       `json:"loss_percent"`
	Sent        int           `json:"sent"`
	Received    int           `json:"received"`
	AvgLatency  time.Duration `json:"avg_latency"`
	MinLatency  time.Duration `json:"min_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	StdDev      time.Duration `json:"std_dev"`
	LastLatency time.Duration `json:"last_latency"`
	Status      HopStatus     `json:"status"`
	ASN         string        `json:"asn,omitempty"`
	Location    string        `json:"location,omitempty"`
}

// HopStatus represents the status of a hop
type HopStatus string

const (
	HopStatusOK       HopStatus = "ok"
	HopStatusTimeout  HopStatus = "timeout"
	HopStatusLoss     HopStatus = "packet_loss"
	HopStatusHighLat  HopStatus = "high_latency"
)

// MTRSummary contains summary information about the trace
type MTRSummary struct {
	DestinationReached bool          `json:"destination_reached"`
	TotalLatency       time.Duration `json:"total_latency"`
	WorstHop           int           `json:"worst_hop"`
	WorstHopHost       string        `json:"worst_hop_host"`
	WorstHopLoss       float64       `json:"worst_hop_loss"`
	WorstHopLatency    time.Duration `json:"worst_hop_latency"`
	Bottleneck         string        `json:"bottleneck,omitempty"`
	Recommendation     string        `json:"recommendation,omitempty"`
}
