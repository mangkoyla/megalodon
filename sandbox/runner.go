package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
)

const connectivityTest = "https://myip.shylook.workers.dev"

func testSingConfig(singConfig option.Options) (configGeoipStruct, error) {
	configGeoip := configGeoipStruct{}
	boxInstance, err := box.New(box.Options{
		Context: context.Background(),
		Options: singConfig,
	})
	if err != nil {
		return configGeoip, err
	}

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetProxy(fmt.Sprintf("socks5://0.0.0.0:%v", singConfig.Inbounds[0].MixedOptions.ListenPort))

	// Start test
	defer boxInstance.Close()
	if err := boxInstance.Start(); err != nil {
		return configGeoip, err
	}

	resp, err := client.R().Get(connectivityTest)
	if err != nil {
		return configGeoip, err
	}

	if resp.StatusCode() == 200 {
		json.Unmarshal(resp.Body(), &configGeoip)
		return configGeoip, nil
	}

	return configGeoip, errors.New(resp.Status())
}
