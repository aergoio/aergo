package context

import (
	"strings"
	"unicode"
)

func ParseFirstWord(input string) (string, string) {

	trimedStr := strings.TrimSpace(input) //remove whitespace
	splitedStr := strings.Fields(trimedStr)
	if len(splitedStr) == 0 {
		return "", ""
	} else if len(splitedStr) == 1 {
		return strings.ToLower(splitedStr[0]), ""
	}

	return strings.ToLower(splitedStr[0]), strings.Join(splitedStr[1:], " ")
}

func ParseAccentString(input string) []string {

	var ret []string

	for _, inputPiece := range strings.Split(input, "`") {
		for _, character := range inputPiece {
			// ignore inputPiece that has only whitespace
			if unicode.IsSpace(character) == false {
				ret = append(ret, inputPiece)

				break
			}
		}
	}

	return ret
}
