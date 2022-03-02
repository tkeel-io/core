package util

func WrapS(str string) string {
	return "\"" + str + "\""
}

func UnwrapS(s string) string {
	quoteChar := byte('"')
	if len(s) > 2 {
		if s[0] == quoteChar && s[len(s)-1] == quoteChar {
			return s[1 : len(s)-1]
		}
	}
	return s
}
