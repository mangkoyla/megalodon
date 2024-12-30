package helper

import "github.com/FoolVPN-ID/Megalodon/common/shared"

func CCToEmoji(cc string) string {
	for _, country := range shared.CountryList {
		if cc == country.Code {
			return string(0x1F1E6+rune(country.Code[0])-'A') + string(0x1F1E6+rune(country.Code[1])-'A')
		}
	}

	return cc
}

func GetRegionFromCC(cc string) string {
	for _, country := range shared.CountryList {
		if country.Code == cc {
			return country.Region
		}
	}

	return ""
}
