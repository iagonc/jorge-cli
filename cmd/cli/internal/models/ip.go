package models

// IPInfo represents information about an IP address
type IPInfo struct {
	IP          string      `json:"ip"`
	Version     string      `json:"version"` // ipv4 or ipv6
	IsPrivate   bool        `json:"is_private"`
	IsLoopback  bool        `json:"is_loopback"`
	Hostname    string      `json:"hostname,omitempty"`
	Geo         *GeoInfo    `json:"geo,omitempty"`
	ASN         *ASNInfo    `json:"asn,omitempty"`
	DNS         *IPDNSInfo  `json:"dns,omitempty"`
	Ports       []PortInfo  `json:"ports,omitempty"`
}

// GeoInfo represents geographic information
type GeoInfo struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

// ASNInfo represents ASN information
type ASNInfo struct {
	Number       int    `json:"number"`
	Organization string `json:"organization"`
	ISP          string `json:"isp"`
}

// IPDNSInfo represents DNS information for IP lookups
type IPDNSInfo struct {
	PTR     []string `json:"ptr,omitempty"`
	Forward []string `json:"forward,omitempty"`
}

// PortInfo represents a port check result
type PortInfo struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Service string `json:"service"`
}
