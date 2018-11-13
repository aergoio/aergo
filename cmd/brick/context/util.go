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

func SplitSpaceAndAccent(input string, addLastInComplete bool) []string {

	var ret []string

	fieldSplit := strings.Fields(input)
	startAccent := -1

	for i, chunk := range fieldSplit {
		if startAccent == -1 {
			if strings.HasPrefix(chunk, "`") {
				if strings.Count(chunk, "`") > 1 && strings.HasSuffix(chunk, "`") {
					// for example `keyword`
					ret = append(ret, chunk[1:len(chunk)-1])
				} else {
					// for example `white space`
					startAccent = i
				}
			} else {
				ret = append(ret, chunk) // just normal keyword
			}
		} else if startAccent != -1 && strings.HasSuffix(chunk, "`") {

			//end of statement
			mergedStr := strings.Join(fieldSplit[startAccent:i+1], " ")
			ret = append(ret, mergedStr[1:len(mergedStr)-1])

			startAccent = -1 // reset
		}

		// contain last incomplete word
		if addLastInComplete && i+1 == len(fieldSplit) && -1 != startAccent {
			mergedStr := strings.Join(fieldSplit[startAccent:], " ")
			ret = append(ret, mergedStr[1:])
		}
	}

	return ret
}
