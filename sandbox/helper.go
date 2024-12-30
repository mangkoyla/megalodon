package sandbox

import (
	"encoding/json"
	"fmt"

	"github.com/sagernet/sing-box/option"
)

func MakeUniqueID(out option.Outbound) string {
	var (
		uniqueID   string
		outMapping = map[string]any{}
		outAny, _  = out.RawOptions()
		outByte, _ = json.Marshal(outAny)
	)

	json.Unmarshal(outByte, &outMapping)

	uniqueID = fmt.Sprintf("%v_%v_%v", outMapping["server"], outMapping["uuid"], outMapping["password"])

	if outMapping["transport"] != nil {
		if outMapping["transport"].(map[string]any)["headers"] != nil {
			uniqueID += fmt.Sprintf("_%v", outMapping["transport"].(map[string]any)["headers"].(map[string]any)["Host"])
		} else {
			uniqueID += fmt.Sprintf("_%v", outMapping["transport"].(map[string]any)["host"])
		}
	}

	if outMapping["tls"] != nil {
		uniqueID += fmt.Sprintf("_%s", outMapping["tls"].(map[string]any)["server_name"])
	}

	return uniqueID
}
