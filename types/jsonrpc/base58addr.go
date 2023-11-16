package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/types"
)

func FillTxBody(source *InOutTxBody, target *types.TxBody) error {
	var err error
	if source == nil {
		return errors.New("tx body is empty")
	}
	target.Nonce = source.Nonce
	if source.Account != "" {
		target.Account, err = types.DecodeAddress(source.Account)
		if err != nil {
			return err
		}
	}
	if source.Recipient != "" {
		target.Recipient, err = types.DecodeAddress(source.Recipient)
		if err != nil {
			return err
		}
	}
	if source.Amount != "" {
		amount, err := ParseUnit(source.Amount)
		if err != nil {
			return err
		}
		target.Amount = amount.Bytes()
	}
	if source.Payload != "" {
		target.Payload, err = base58.Decode(source.Payload)
		if err != nil {
			return err
		}
	}
	target.GasLimit = source.GasLimit
	if source.GasPrice != "" {
		price, err := ParseUnit(source.GasPrice)
		if err != nil {
			return err
		}
		target.GasPrice = price.Bytes()
	}
	if source.ChainIdHash != "" {
		target.ChainIdHash, err = base58.Decode(source.ChainIdHash)
		if err != nil {
			return err
		}
	}
	if source.Sign != "" {
		target.Sign, err = base58.Decode(source.Sign)
		if err != nil {
			return err
		}
	}
	target.Type = source.Type
	return nil
}

func ParseBase58Tx(jsonTx []byte) ([]*types.Tx, error) {
	var inputlist []InOutTx
	err := json.Unmarshal([]byte(jsonTx), &inputlist)
	if err != nil {
		var input InOutTx
		err = json.Unmarshal([]byte(jsonTx), &input)
		if err != nil {
			return nil, err
		}
		inputlist = append(inputlist, input)
	}
	txs := make([]*types.Tx, len(inputlist))
	for i, in := range inputlist {
		tx := &types.Tx{Body: &types.TxBody{}}
		if in.Hash != "" {
			tx.Hash, err = base58.Decode(in.Hash)
			if err != nil {
				return nil, err
			}
		}
		err = FillTxBody(in.Body, tx.Body)
		if err != nil {
			return nil, err
		}
		txs[i] = tx
	}

	return txs, nil
}

func ParseBase58TxBody(jsonTx []byte) (*types.TxBody, error) {
	body := &types.TxBody{}
	in := &InOutTxBody{}

	err := json.Unmarshal(jsonTx, in)
	if err != nil {
		return nil, err
	}

	err = FillTxBody(in, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func ConvPeerLong(p *types.Peer) *LongInOutPeer {
	out := &LongInOutPeer{InOutPeer: *ConvPeer(p)}
	out.ProducerIDs = make([]string, len(p.Address.ProducerIDs))
	for i, pid := range p.Address.ProducerIDs {
		out.ProducerIDs[i] = base58.Encode(pid)
	}
	if p.Address.Role == types.PeerRole_Agent {
		out.Certificates = make([]*InOutCert, len(p.Certificates))
		for i, cert := range p.Certificates {
			addrs := []string{}
			for _, ad := range cert.AgentAddress {
				addrs = append(addrs, string(ad))
			}
			out.Certificates[i] = &InOutCert{CertVersion: cert.CertVersion,
				ProducerID: base58.Encode(cert.BPID), AgentID: base58.Encode(cert.AgentID),
				CreateTime: time.Unix(0, cert.CreateTime), ExpireTime: time.Unix(0, cert.ExpireTime),
				Addresses: addrs}
		}
	}
	return out
}

func BlockConvBase58Addr(b *types.Block) string {
	return toString(ConvBlock(b))
}

func PeerListToString(p *types.PeerList) string {
	peers := []*InOutPeer{}
	for _, peer := range p.GetPeers() {
		peers = append(peers, ConvPeer(peer))
	}
	return toString(peers)
}
func ShortPeerListToString(p *types.PeerList) string {
	var peers []string
	for _, peer := range p.GetPeers() {
		pa := peer.Address
		peers = append(peers, fmt.Sprintf("%s;%s/%d;%s;%d", p2putil.ShortForm(types.PeerID(pa.PeerID)), pa.Address, pa.Port, peer.AcceptedRole.String(), peer.Bestblock.BlockNo))
	}
	return toString(peers)
}
func LongPeerListToString(p *types.PeerList) string {
	peers := []*LongInOutPeer{}
	for _, peer := range p.GetPeers() {
		peers = append(peers, ConvPeerLong(peer))
	}
	return toString(peers)
}
func toString(out interface{}) string {
	jsonout, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return ""
	}
	return string(jsonout)
}
