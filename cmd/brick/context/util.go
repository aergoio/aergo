package context

import (
	"strings"
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

type Chunk struct {
	Accent bool
	Text   string
}

func SplitSpaceAndAccent(input string, addLastInComplete bool) []Chunk {

	var ret []Chunk

	fieldSplit := strings.Fields(input)
	startAccent := -1

	for i, chunk := range fieldSplit {
		if startAccent == -1 {
			if strings.HasPrefix(chunk, "`") {
				if strings.Count(chunk, "`") > 1 && strings.HasSuffix(chunk, "`") {
					// for example `keyword`
					ret = append(ret, Chunk{Accent: true, Text: chunk[1 : len(chunk)-1]})
				} else {
					// for example `white space`
					startAccent = i
				}
			} else {
				// just normal keyword
				ret = append(ret, Chunk{Accent: false, Text: chunk})
			}
		} else if startAccent != -1 && strings.HasSuffix(chunk, "`") {

			//end of statement
			mergedStr := strings.Join(fieldSplit[startAccent:i+1], " ")
			ret = append(ret, Chunk{Accent: true, Text: mergedStr[1 : len(mergedStr)-1]})

			startAccent = -1 // reset
		}

		// contain last incomplete word
		if addLastInComplete && i+1 == len(fieldSplit) && -1 != startAccent {
			mergedStr := strings.Join(fieldSplit[startAccent:], " ")
			ret = append(ret, Chunk{Accent: true, Text: mergedStr[1:]})
		}
	}

	return ret
}
