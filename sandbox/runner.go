package sandbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FoolVPN-ID/Megalodon/common/helper"
	"github.com/Noooste/azuretls-client"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
)

const connectivityTest = "https://myip.shylook.workers.dev"

func testSingConfigWithContext(singConfig option.Options, ctx context.Context) (configGeoipStruct, error) {
	// Re-allocate free port
	freePort := helper.GetFreePort()
	singConfig.Inbounds[0].MixedOptions.ListenPort = uint16(freePort)

	configGeoip := configGeoipStruct{}
	boxInstance, err := box.New(box.Options{
		Context: ctx,
		Options: singConfig,
	})
	if err != nil {
		return configGeoip, err
	}

	// Start sing-box
	defer boxInstance.Close()
	if err := boxInstance.Start(); err != nil {
		return configGeoip, err
	}

	session := azuretls.NewSessionWithContext(ctx)
	session.InsecureSkipVerify = true
	defer session.Close()

	session.SetProxy(fmt.Sprintf("socks5://0.0.0.0:%v", freePort))

	if err := session.Connect(connectivityTest); err != nil {
		return configGeoip, err
	}

	resp, err := session.Get(connectivityTest)
	if err != nil {
		return configGeoip, err
	} else {
		if resp.StatusCode == 200 {
			json.Unmarshal(resp.Body, &configGeoip)
		}
	}

	return configGeoip, nil
}
