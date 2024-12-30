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

func testSingConfigWithContext(singConfig option.Options, ctx context.Context) (configGeoipStruct, error) {
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

	var (
		errChan = make(chan error)
		isDone  = make(chan int)
	)
	go func() {
		// Resty seems problematic, subject to be changed
		client := resty.New()
		client.SetTimeout(5 * time.Second)
		client.SetProxy(fmt.Sprintf("socks5://0.0.0.0:%v", singConfig.Inbounds[0].MixedOptions.ListenPort))

		resp, err := client.R().Get(connectivityTest)
		if err != nil {
			errChan <- err
		} else {
			if resp.StatusCode() == 200 {
				json.Unmarshal(resp.Body(), &configGeoip)
			}
		}

		close(errChan)
		isDone <- 1
	}()

	select {
	case <-ctx.Done():
		return configGeoip, errors.New("operation timeout")
	case <-isDone:
		for err := range errChan {
			if err != nil {
				return configGeoip, err
			}
		}
		return configGeoip, nil
	}
}
