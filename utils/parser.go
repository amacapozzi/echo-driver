package utils

import "fmt"

func ParseError(err error) {
	fmt.Println("[=] Error:", err.Error())
}

func ParseMemory(memory []byte) []string {
	var result []string
	var currentStr []byte

	for _, b := range memory {
		if b >= 32 && b <= 126 {
			currentStr = append(currentStr, b)
		} else if len(currentStr) > 0 {
			result = append(result, string(currentStr))
			currentStr = currentStr[:0]
		}
	}

	if len(currentStr) > 0 {
		result = append(result, string(currentStr))
	}

	return result
}
