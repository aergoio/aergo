// +build Debug

package contract

import (
	"encoding/binary"
)

func getCompiledABI(code string) ([]byte, error) {
	byteCode, err := compile(code)
	if err != nil {
		return nil, err
	}
	codeLen := binary.LittleEndian.Uint32(byteCode[:4])
	return byteCode[4+codeLen:], nil
}

func NewLuaTxDefBig(sender, contract string, amount *big.Int, code string) *luaTxDef {
	byteAbi, err := getCompiledABI(code)
	if err != nil {
		return &luaTxDef{cErr: err}
	}
	byteCode := []byte(code)
	payload := make([]byte, 8+len(byteCode)+len(byteAbi))
	binary.LittleEndian.PutUint32(payload[0:], uint32(len(byteCode)+len(byteAbi)+8))
	binary.LittleEndian.PutUint32(payload[4:], uint32(len(byteCode)))
	codeLen := copy(payload[8:], byteCode)
	copy(payload[8+codeLen:], byteAbi)

	return &luaTxDef{
		luaTxCommon: luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			code:     payload,
			amount:   amount,
			id:       newTxId(),
		},
		cErr: nil,
	}
}

