package context

import (
	"fmt"
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

func ParseDecimalAmount(str string, digits int) string {

	// if there is no 'aergo' unit, just return the amount str
	if !strings.HasSuffix(strings.ToLower(str), "aergo") {
		return str
	}

	// remove the 'aergo' unit
	str = str[:len(str)-5]
	// trim trailing spaces
	str = strings.TrimRight(str, " ")

	// get the position of the decimal point
	idx := strings.Index(str, ".")

	// if not found, just add the leading zeros
	if idx == -1 {
		return str + strings.Repeat("0", digits)
	}

	// Get the integer and decimal parts
	p1 := str[0:idx]
	p2 := str[idx+1:]

	// Check for another decimal point
	if strings.Index(p2, ".") != -1 {
		return "error"
	}

	// Compute the amount of zero digits to add
	to_add := digits - len(p2)
	if to_add > 0 {
		p2 = p2 + strings.Repeat("0", to_add)
	} else if to_add < 0 {
		// Do not truncate decimal amounts
		return "error"
	}

	// Join the integer and decimal parts
	str = p1 + p2

	// Remove leading zeros
	str = strings.TrimLeft(str, "0")
	if str == "" {
		str = "0"
	}
	return str
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

func IsCompleteCommand(line string, line_no int, isOpen bool) (bool, error) {

	chunks := strings.Fields(line)

	for _, chunk := range chunks {
		if isOpen {
			if chunk == "`" {
				isOpen = false
			} else if strings.HasPrefix(chunk, "`") {
				return false, fmt.Errorf("already open parameter at line %v", line_no)
			} else if strings.HasSuffix(chunk, "`") {
				isOpen = false
			}
		} else {
			if chunk == "`" {
				isOpen = true
			} else if strings.HasPrefix(chunk, "`") {
				if strings.HasSuffix(chunk, "`") {
					// for example `keyword`
				} else {
					// for example `white space`
					isOpen = true
				}
			} else if strings.HasSuffix(chunk, "`") {
				return false, fmt.Errorf("closing not open parameter at line %v", line_no)
			}
		}
	}

	return isOpen, nil
}
