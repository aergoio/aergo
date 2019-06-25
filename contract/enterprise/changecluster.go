package enterprise

import (
	"encoding/binary"
	"fmt"
	"github.com/aergoio/aergo/types"
	"reflect"
	"strconv"
)

const (
	CmdMembershipAdd    = "add"
	CmdMembershipRemove = "remove"

	CCCommand        = "command"
	MemberAttrName   = "name"
	MemberAttrUrl    = "url"
	MemberAttrPeerID = "peerid"
	MemberAttrID     = "id"
)

func validateChangeCluster(ci types.CallInfo, txHash []byte) (interface{}, error) {
	var (
		ccArg     ccArgument
		ok        bool
		err       error
		changeReq *types.MembershipChange
	)

	if len(ci.Args) != 1 { //args[0] : map{ "command": "add", "name:", "url:", "peerid:", "id:"}
		return nil, fmt.Errorf("invalid arguments in payload for ChangeCluster: %s", ci.Args)
	}

	ccArg, ok = ci.Args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid argument in payload for ChangeCluster(map[string]interface{}) : argument=%v", reflect.TypeOf(ci.Args[0]))
	}

	if changeReq, err = ccArg.parse(); err != nil {
		return nil, err
	}

	changeReq.RequestID = generateCCIDFromTxID(txHash)

	return changeReq, nil
}

func generateCCIDFromTxID(txHash []byte) uint64 {
	return binary.LittleEndian.Uint64(txHash[:8])
}

func (cc ccArgument) get(key string) (string, error) {
	var (
		val    interface{}
		valStr string
		ok     bool
	)

	if val, ok = cc[key]; !ok {
		return "", fmt.Errorf("invalid ChangeCluster argument: not exist %s", key)
	}

	if valStr, ok = val.(string); !ok {
		return "", fmt.Errorf("invalid ChangeCluster argument: not string value(%v)", reflect.TypeOf(val))
	}

	return valStr, nil
}

func (cc ccArgument) getUint64(key string) (uint64, error) {
	var (
		val     interface{}
		valUint float64
		ok      bool
	)

	if val, ok = cc[key]; !ok {
		return 0, fmt.Errorf("invalid ChangeCluster argument: not exist %s", key)
	}

	if valUint, ok = val.(float64); !ok {
		return 0, fmt.Errorf("invalid ChangeCluster argument: not uint64(%v)", reflect.TypeOf(val))
	}

	return uint64(valUint), nil
}

func (cc ccArgument) parse() (*types.MembershipChange, error) {
	var (
		cmd, name, url, peerid, idStr string
		id                            uint64
		err                           error
		mChange                       types.MembershipChange
	)

	if cmd, err = cc.get(CCCommand); err != nil {
		return nil, err
	}

	switch cmd {
	case CmdMembershipAdd:
		mChange.Type = types.MembershipChangeType_ADD_MEMBER

		if name, err = cc.get(MemberAttrName); err != nil {
			return nil, err
		}

		if url, err = cc.get(MemberAttrUrl); err != nil {
			return nil, err
		}

		if peerid, err = cc.get(MemberAttrPeerID); err != nil {
			return nil, err
		}

		_, err = types.IDB58Decode(peerid)
		if err != nil {
			return nil, fmt.Errorf("invalid ChangeCluster argument: invalid peerid %s", peerid)
		}

	case CmdMembershipRemove:
		mChange.Type = types.MembershipChangeType_REMOVE_MEMBER

		if idStr, err = cc.get(MemberAttrID); err != nil {
			return nil, err
		}

		if id, err = strconv.ParseUint(idStr, 16, 64); err != nil {
			return nil, fmt.Errorf("invalid ChangeCluster argument: invalid id %s. ID must be a string in hexadecial format(ex:dd44cf1a06727dc5)", idStr)
		}

	default:
		return nil, fmt.Errorf("invalid ChangeCluster argument: invalid command %s", cmd)
	}

	mChange.Attr = &types.MemberAttr{Name: name, Url: url, PeerID: []byte(peerid), ID: id}

	return &mChange, nil
}
