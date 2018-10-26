/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

func calculateFieldDescSize(varSize int) int{
	switch {
	case varSize == 0 :
		return 0
	case varSize < 128 :
		return 2
	case varSize < 16384 :
		return 3
	case varSize < 2097152 :
		return 4
	case varSize < 268435456:
		return 5
	case varSize < 34359738368:
		return 6
	default:
		return 7
	}
}
