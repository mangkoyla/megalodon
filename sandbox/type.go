package sandbox

import "github.com/sagernet/sing-box/option"

type configGeoipStruct struct {
	IP             string `json:"ip"`
	Proxy          string `json:"proxy"`
	Port           int64  `json:"port"`
	Country        string `json:"country"`
	AsOrganization string `json:"asOrganization"`
}

type TestResultStruct struct {
	TestPassed  []string
	ConfigGeoip configGeoipStruct
	Outbound    option.Outbound
	RawConfig   string
}
