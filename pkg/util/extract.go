package util

func ExtractMap(m map[string]string) string {
	var result string
	for key, value := range m {
		result += "," + key + "=" + value
	}
	return result[1:]
}
