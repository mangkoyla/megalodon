package helper

func RemoveEmptyStringFromList(l []string) []string {
	results := []string{}
	for _, v := range l {
		if v != "" {
			results = append(results, v)
		}
	}

	return results
}
