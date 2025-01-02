package provider

import (
	"sync"

	logger "github.com/FoolVPN-ID/megalodon/log"
)

type providerStruct struct {
	subs   []providerSubStruct
	Nodes  []string
	logger logger.LoggerStruct
	sync.Mutex
}

func MakeSubProvider() *providerStruct {
	prov := providerStruct{
		logger: *logger.MakeLogger(),
	}

	return &prov
}
