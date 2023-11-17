package jsonrpc

type EncodingType int

const (
	Raw EncodingType = 0 + iota
	Base58
)

func (b *InOutTxBody) String() string {
	return B58JSON(b)
}

func (t *InOutTx) String() string {
	return B58JSON(t)
}

func (t *InOutTxInBlock) String() string {
	return B58JSON(t)
}
