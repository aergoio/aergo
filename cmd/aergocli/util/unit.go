package util

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

//var unit map[string]*big.Int
var units map[string]int
var unitlist []string

func init() {
	units = map[string]int{}
	units["aergo"] = 18
	units["gaer"] = 9
	units["aer"] = 0
	unitlist = []string{"aergo", "gaer", "aer"}
}

func ParseUnit(s string) (*big.Int, error) {
	result, ok := new(big.Int).SetString(strings.TrimSpace(s), 10)
	if !ok {
		lower := strings.ToLower(s)
		for _, v := range unitlist {
			if strings.Contains(lower, v) {
				number := strings.TrimSpace(strings.TrimSuffix(lower, v))
				numbers := strings.Split(number, ".")
				var nstr string
				var pos int
				switch len(numbers) {
				case 1:
					nstr = numbers[0]
					pos = 0
				case 2:
					if len(numbers[1]) > units[v] {
						return big.NewInt(0), fmt.Errorf("too small unit %s", s)
					}
					nstr = numbers[0] + numbers[1]
					pos = len(numbers[1])
				default:
					continue
				}
				for i := pos; i < units[v]; i++ {
					nstr += "0"
				}
				n, ok := new(big.Int).SetString(nstr, 10)
				if !ok {
					continue
				} else {
					return n, nil
				}
			}
		}
		return big.NewInt(0), fmt.Errorf("could not parse %s", s)
	}
	return result, nil
}

func ConvertUnit(n *big.Int, unit string) (string, error) {
	unit = strings.ToLower(unit)
	nstr := n.String()

	if len(nstr) > units[unit] {
		dotpos := len(nstr) - units[unit]
		nstr = nstr[0:dotpos] + "." + nstr[dotpos:]
		return strings.TrimRight(strings.TrimRight(nstr, "0"), ".") + " " + unit, nil
	}
	result := strings.TrimRight(fmt.Sprintf(".%0"+strconv.Itoa(units[unit])+"d", n), ".0")
	result = "0" + result + " " + unit
	return result, nil
}
