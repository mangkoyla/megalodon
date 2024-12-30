package provider

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/FoolVPN-ID/Megalodon/common/helper"
	"github.com/FoolVPN-ID/Megalodon/constant"
	"github.com/go-resty/resty/v2"
)

var configSeparators = []string{"\n", "|", ",", "<br/>"}

func (prov *providerStruct) GatherSubFile() {
	var subFileUrlString, err = helper.ReadFileAsString("./resources/sublist.json")
	var subFileUrls = []string{}

	if err != nil {
		prov.logger.Error(err.Error())
		return
	}

	json.Unmarshal([]byte(subFileUrlString), &subFileUrls)

	for _, subFileUrl := range subFileUrls {
		client := resty.New()
		resp, err := client.R().Get(subFileUrl)
		if err != nil {
			prov.logger.Error(err.Error())
			continue
		}

		if resp.StatusCode() == 200 {
			var subFile = []providerSubStruct{}
			if err := json.Unmarshal(resp.Body(), &subFile); err == nil {
				prov.subs = append(prov.subs, subFile...)
			}
		}
	}
}

func (prov *providerStruct) GatherNodes() {
	var (
		wg    = sync.WaitGroup{}
		queue = make(chan struct{}, 10)
	)

	client := resty.New()
	client.SetTimeout(10 * time.Second)
	for i, sub := range prov.subs {
		var subUrls = strings.Split(sub.URL, "|")
		for x, subUrl := range subUrls {
			wg.Add(1)
			queue <- struct{}{}

			prov.logger.Info(fmt.Sprintf("[[%d/%d]%d/%d] [%d] %s\n", x, len(subUrls), i, len(prov.subs), len(prov.Nodes), subUrl))
			go (func() {
				defer func() {
					wg.Done()
					<-queue
				}()
				defer func() {
					recover()
				}()

				resp, err := client.R().Get(subUrl)
				if err != nil {
					panic(err)
				}

				if resp.StatusCode() == 200 {
					nodes := []string{}
					// re-Filter nodes due to some bullshit
					for _, separator := range configSeparators {
						if len(nodes) == 0 {
							nodes = append(nodes, strings.Split(helper.DecodeBase64Safe(resp.String()), separator)...)
						} else {
							filteredNodes := []string{}
							for _, node := range nodes {
								filteredNodes = append(filteredNodes, strings.Split(node, separator)...)
							}

							nodes = filteredNodes
						}
					}

					for _, node := range nodes {
						for _, acceptedType := range constant.ACCEPTED_TYPES {
							if strings.HasPrefix(node, acceptedType) {
								prov.Nodes = append(prov.Nodes, node)
							}
						}
					}
				}
			})()
		}
	}

	// Wait for all goroutines
	wg.Wait()
}
