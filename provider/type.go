package provider

type providerSubStruct struct {
	ID           int    `json:"id"`
	Remarks      string `json:"remarks"`
	Site         string `json:"site"`
	URL          string `json:"url"`
	UpdateMethod string `json:"update_method"`
	Enabled      bool   `json:"enabled"`
}
