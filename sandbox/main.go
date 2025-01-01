package sandbox

import (
	"context"
	"encoding/json"
	"time"

	logger "github.com/FoolVPN-ID/megalodon/log"
	"github.com/FoolVPN-ID/tool/modules/config"
	"github.com/sagernet/sing-box/option"
)

var (
	testTypes = []string{"cdn", "sni"}
	sniHost   = "meet.google.com"
	cdnHost   = "172.67.73.39"
)

type sandboxStruct struct {
	Results []TestResultStruct
	log     *logger.LoggerStruct
}

func MakeSandbox() *sandboxStruct {
	return &sandboxStruct{
		log: logger.MakeLogger(),
	}
}

func (sb *sandboxStruct) TestConfig(rawConfig string, accountIndex, accountTotal int) error {
	singConfig, err := config.BuildSingboxConfig(rawConfig)
	if err != nil {
		return err
	}

	testResult := TestResultStruct{
		Outbound: singConfig.Outbounds[0],
	}

	for _, testType := range testTypes {
		singConfigMapping := map[string]any{}
		singConfigByte, _ := json.Marshal(singConfig)
		json.Unmarshal(singConfigByte, &singConfigMapping)

		outbound := singConfigMapping["outbounds"].([]any)[0].(map[string]any)

		if testType == "cdn" {
			outbound["server"] = cdnHost
		} else {
			if outbound["tls"] != nil {
				outboundTLS := outbound["tls"].(map[string]any)
				if outboundTLS["enabled"] == true {
					outboundTLS["insecure"] = true
					outboundTLS["server_name"] = sniHost

					outbound["tls"] = outboundTLS
				}
			}

			if outbound["transport"] != nil {
				outboundTransport := outbound["transport"].(map[string]any)
				if outboundTransport["headers"] != nil {
					transportHeaders := outboundTransport["headers"].(map[string]any)
					if transportHeaders["Host"] != nil {
						transportHeaders["Host"] = sniHost
					}
					outboundTransport["headers"] = transportHeaders
				}
				if outboundTransport["host"] != nil {
					outboundTransport["host"] = sniHost
				}

				outbound["transport"] = outboundTransport
			}
		}

		singConfigMapping["outbounds"].([]any)[0] = outbound

		configForTest := option.Options{}
		configForTestByte, err := json.Marshal(singConfigMapping)
		if err != nil {
			return err
		}
		configForTest.UnmarshalJSON(configForTestByte)

		// Close closure variable
		func(connMode string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if configGeoip, err := testSingConfigWithContext(configForTest, ctx); err == nil {
				testResult.TestPassed = append(testResult.TestPassed, connMode)
				testResult.ConfigGeoip = configGeoip

				// sb.log.Success(fmt.Sprintf("[%d/%d] [%d+%d] %v %s %s", accountIndex, accountTotal, len(sb.Results), len(testResult.TestPassed), testResult.TestPassed, configGeoip.Country, configGeoip.AsOrganization))
			} else {
				// sb.log.Error(fmt.Sprintf("[%d/%d] %s", accountIndex, accountTotal, err.Error()))
			}
		}(testType)
	}

	if len(testResult.TestPassed) > 0 {
		sb.Results = append(sb.Results, testResult)
	}

	return nil
}
