package types

var quirkTxMap map[TxID][]byte

const (
	// sending aergo to wrong recipient
	B23994084_001 = "85QfFtU62ihqj3LnoiKpZgfpKg8mixKRJ5EBMCHKVPGN"
)

func init() {
	quirkTxMap = make(map[TxID][]byte)
	putTxID(B23994084_001)
}

func putTxID(b58encoded string) {
	hash := DecodeB58(b58encoded)
	id, _ := ParseToTxID(hash)
	quirkTxMap[id] = hash
}

// IsQuirkTx checks if the tx should be handled specially.
func IsQuirkTx(txHash []byte) bool {
	id, _ := ParseToTxID(txHash)
	_, exist := quirkTxMap[id]
	return exist
}
