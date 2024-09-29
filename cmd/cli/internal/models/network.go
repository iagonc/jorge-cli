package models

type DNSRecord struct {
    Type string
    IP   string
}

type DNSLookupResult struct {
    Records []DNSRecord
}

type NSLookupResult struct {
    IP string
}

type TracerouteHop struct {
    HopNumber    int
    Address      string
    ResponseTime string
}

type TracerouteResult struct {
    Hops []TracerouteHop
}

type HTTPRequestResult struct {
    Status       string
    ResponseTime string
    ContentType  string
}

type PingResult struct {
    Sent        int
    Received    int
    Lost        int
    LossPercent float64
    AvgLatency  int
}

type NetstatConnection struct {
    Protocol      string
    LocalAddress  string
    RemoteAddress string
    Status        string
}

type NetstatResult struct {
    Connections []NetstatConnection
}

type IftopConnection struct {
    Source        string
    Destination   string
    SentKBps      string
    ReceivedKBps  string
}

type IftopResult struct {
    SendingKBps    string
    ReceivingKBps  string
    TopConnections []IftopConnection
}

type NetworkDebugResult struct {
    DNSLookup   DNSLookupResult
    NSLookup    NSLookupResult
    Traceroute  TracerouteResult
    HTTPRequest HTTPRequestResult
    Ping        PingResult
    Netstat     NetstatResult
    Iftop       IftopResult
}
