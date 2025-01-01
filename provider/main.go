package provider

import logger "github.com/FoolVPN-ID/megalodon/log"

type providerStruct struct {
	subs   []providerSubStruct
	Nodes  []string
	logger logger.LoggerStruct
}

func MakeSubProvider() *providerStruct {
	prov := providerStruct{
		logger: *logger.MakeLogger(),
	}

	return &prov
}
