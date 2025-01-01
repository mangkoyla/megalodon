package database

type DatabaseFieldStruct struct {
	// VPN fields
	Server      string `json:"server,omitempty"`       // 0
	Ip          string `json:"ip,omitempty"`           // 1
	ServerPort  int    `json:"server_port,omitempty"`  // 2
	UUID        string `json:"uuid,omitempty"`         // 3
	Password    string `json:"password,omitempty"`     // 4
	Security    string `json:"security,omitempty"`     // 5
	AlterId     int    `json:"alter_id,omitempty"`     // 6
	Method      string `json:"method,omitempty"`       // 7
	Plugin      string `json:"plugin,omitempty"`       // 8
	PluginOpts  string `json:"plugin_opts,omitempty"`  // 9
	Host        string `json:"host,omitempty"`         // 10
	TLS         bool   `json:"tls,omitempty"`          // 11
	Transport   string `json:"transport,omitempty"`    // 12
	Path        string `json:"path,omitempty"`         // 13
	ServiceName string `json:"service_name,omitempty"` // 14
	Insecure    bool   `json:"insecure,omitempty"`     // 15
	SNI         string `json:"sni,omitempty"`          // 16

	// Service fields
	Remark      string `json:"remark,omitempty"`       // 17
	ConnMode    string `json:"conn_mode,omitempty"`    // 18
	CountryCode string `json:"country_code,omitempty"` // 19
	Region      string `json:"region,omitempty"`       // 20
	Org         string `json:"org,omitempty"`          // 21
	VPN         string `json:"vpn,omitempty"`          // 22

	// Additional fields
	Raw string `json:"raw"` // 23
}
