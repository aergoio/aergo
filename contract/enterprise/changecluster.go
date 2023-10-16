package enterprise

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/aergoio/aergo/v2/types"
)

type CcArgument map[string]interface{}

const (
	CmdMembershipAdd    = "add"
	CmdMembershipRemove = "remove"

	CCCommand         = "command"
	MemberAttrName    = "name"
	MemberAttrAddress = "address"
	MemberAttrPeerID  = "peerid"
	MemberAttrID      = "id"
)

/*
var (
	ConfChangeState_name = map[ConfChangeState]string{
		0: "Proposed",
		1: "Saved",
		2: "Applied",
	}
)*/

func ValidateChangeCluster(ci types.CallInfo, blockNo types.BlockNo) (interface{}, error) {
	var (
		ccArg     CcArgument
		ok        bool
		err       error
		changeReq *types.MembershipChange
	)

	if len(ci.Args) != 1 { //args[0] : map{ "command": "add", "name:", "address:", "peerid:", "id:"}
		return nil, fmt.Errorf("invalid arguments in payload for ChangeCluster: %s", ci.Args)
	}

	ccArg, ok = ci.Args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid argument in payload for ChangeCluster(map[string]interface{}) : argument=%v", reflect.TypeOf(ci.Args[0]))
	}

	if changeReq, err = ccArg.parse(); err != nil {
		return nil, err
	}

	changeReq.RequestID = blockNo

	return changeReq, nil
}

func (cc CcArgument) get(key string) (string, error) {
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

func (cc CcArgument) getUint64(key string) (uint64, error) {
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

func (cc CcArgument) parse() (*types.MembershipChange, error) {
	var (
		cmd, name, address, peeridStr, idStr string
		id                                   uint64
		err                                  error
		mChange                              types.MembershipChange
		peerID                               types.PeerID
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

		if address, err = cc.get(MemberAttrAddress); err != nil {
			return nil, err
		}

		if peeridStr, err = cc.get(MemberAttrPeerID); err != nil {
			return nil, err
		}

		peerID, err = types.IDB58Decode(peeridStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ChangeCluster argument: can't decode peerid string(%s)", peeridStr)
		}

		if _, err := types.ParseMultiaddr(address); err != nil {
			return nil, fmt.Errorf("invalid ChangeCluster argument: %s", err.Error())
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

	mChange.Attr = &types.MemberAttr{Name: name, Address: address, PeerID: []byte(peerID), ID: id}

	return &mChange, nil
}
