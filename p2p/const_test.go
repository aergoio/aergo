/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// this file collect sample global constants used in unit test. I'm tired of creating less meaningful variables in each tests.

var dummyBlockHash, _ = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
var dummyBlockHeight uint64 = 100215
var dummyTxHash, _ = enc.ToBytes("4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45")

var samplePeerID types.PeerID
var sampleMeta p2pcommon.PeerMeta
var sampleErr error

var logger *log.Logger

func init() {
	logger = log.NewLogger("test")
	samplePeerID, _ = types.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	sampleErr = fmt.Errorf("err in unittest")
	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta = p2pcommon.PeerMeta{ID: samplePeerID, Addresses: []types.Multiaddr{sampleMA}}

}

const (
	sampleKey1PrivBase64 = "CAISIM1yE7XjJyKTw4fQYMROnlxmEBut5OPPGVde7PeVAf0x"
	sampelKey1PubBase64  = "CAISIQOMA3AHgprpAb7goiDGLI6b/je3JKiYSBHyb46twYV7RA=="
	sampleKey1IDbase58   = "16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD"

	sampleKey2PrivBase64 = "CAISIPK7nwYwZwl0LVvjJjZ58gN/4z0iAIOEOi5fKDIMBKCN"
)

var sampleMsgID p2pcommon.MsgID
var sampleKey1Priv crypto.PrivKey
var sampleKey1Pub crypto.PubKey
var sampleKey1ID types.PeerID

var sampleKey2Priv crypto.PrivKey
var sampleKey2Pub crypto.PubKey
var sampleKey2ID types.PeerID

var dummyPeerID types.PeerID
var dummyPeerID2 types.PeerID
var dummyPeerID3 types.PeerID

var dummyBestBlock *types.Block
var dummyMeta p2pcommon.PeerMeta

func init() {
	bytes, _ := base64.StdEncoding.DecodeString(sampleKey1PrivBase64)
	sampleKey1Priv, _ = crypto.UnmarshalPrivateKey(bytes)
	bytes, _ = base64.StdEncoding.DecodeString(sampelKey1PubBase64)
	sampleKey1Pub, _ = crypto.UnmarshalPublicKey(bytes)
	if !sampleKey1Priv.GetPublic().Equals(sampleKey1Pub) {
		panic("problem in pk generation ")
	}
	sampleKey1ID, _ = types.IDFromPublicKey(sampleKey1Pub)
	if sampleKey1ID.Pretty() != sampleKey1IDbase58 {
		panic("problem in id generation")
	}

	bytes, _ = base64.StdEncoding.DecodeString(sampleKey2PrivBase64)
	sampleKey2Priv, _ = crypto.UnmarshalPrivateKey(bytes)
	sampleKey2Pub = sampleKey2Priv.GetPublic()
	sampleKey2ID, _ = types.IDFromPublicKey(sampleKey2Pub)

	sampleMsgID = p2pcommon.NewMsgID()

	dummyPeerID = sampleKey1ID
	dummyPeerID2, _ = types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	dummyPeerID3, _ = types.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")

	dummyBestBlock = &types.Block{Header: &types.BlockHeader{}}
	dummyMeta = p2pcommon.PeerMeta{ID: dummyPeerID}
}

var sampleTxsB58 = []string{
	"3vMrViqzeBJ9RQ9S6RnSn7dMdoucWVTqNGBLkZkqDLLQ",
	"7wBPhtoJYLrUsu6FWnp4xLk8iU4tczAQiafpnRBmiaSn",
	"H7kzE7S9NEG6Y2xDtfZXTSZHu6E9YUPeeWApqU1x5BaX",
	"7Cahs398NJPkuckeaeChw4DJYCarwq1vMQ2qsYRZMNbn",
	"BxKmDg9VbWHxrWnStEeTzJ2Ze7RF7YK4rpyjcsWSsnxs",
	"DwmGqFU4WgADpYN36FXKsYxMjeppvh9Najg4KxJ8gtX3",
}

var sampleTxs [][]byte
var sampleTxIDs []types.TxID

var sampleBlksB58 = []string{
	"v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6",
	"2VEPg4MqJUoaS3EhZ6WWSAUuFSuD4oSJ645kSQsGV7H9",
	"AtzTZ2CZS45F1276RpTdLfYu2DLgRcd9HL3aLqDT1qte",
	"2n9QWNDoUvML756X7xdHWCFLZrM4CQEtnVH2RzG5FYAw",
	"6cy7U7XKYtDTMnF3jNkcJvJN5Rn85771NSKjc5Tfo2DM",
	"3bmB8D37XZr4DNPs64NiGRa2Vw3i8VEgEy6Xc2XBmRXC",
}
var sampleBlks [][]byte
var sampleBlksIDs []types.BlockID

func init() {
	sampleTxs = make([][]byte, len(sampleTxsB58))
	sampleTxIDs = make([]types.TxID, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
		copy(sampleTxIDs[i][:], hash)
	}

	sampleBlks = make([][]byte, len(sampleBlksB58))
	sampleBlksIDs = make([]types.BlockID, len(sampleBlksB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleBlks[i] = hash
		copy(sampleBlksIDs[i][:], hash)
	}
}

var dummyMo *p2pmock.MockMsgOrder

func createDummyMo(ctrl *gomock.Controller) *p2pmock.MockMsgOrder {
	dummyMo = p2pmock.NewMockMsgOrder(ctrl)
	dummyMo.EXPECT().IsNeedSign().Return(true).AnyTimes()
	dummyMo.EXPECT().IsRequest().Return(true).AnyTimes()
	dummyMo.EXPECT().GetProtocolID().Return(p2pcommon.NewTxNotice).AnyTimes()
	dummyMo.EXPECT().GetMsgID().Return(p2pcommon.NewMsgID()).AnyTimes()
	return dummyMo
}

func TestTXIDs(t *testing.T) {
	for i := 0; i < len(sampleTxIDs); i++ {
		if !bytes.Equal(sampleTxs[i], sampleTxIDs[i][:]) {
			t.Errorf("TX hash %v and converted ID %v are differ.", enc.ToString(sampleTxs[i]), sampleTxIDs[i])
		}
	}
}

func TestBlockIDs(t *testing.T) {
	for i := 0; i < len(sampleBlksIDs); i++ {
		if !bytes.Equal(sampleBlks[i], sampleBlksIDs[i][:]) {
			t.Errorf("TX hash %v and converted ID %v are differ.", enc.ToString(sampleBlks[i]), sampleBlksIDs[i])
		}
	}
}

var testIn = []string{
	"3vMrViqzeBJ9RQ9S6RnSn7dMdoucWVTqNGBLkZkqDLLQ",
	"7wBPhtoJYLrUsu6FWnp4xLk8iU4tczAQiafpnRBmiaSn",
	"H7kzE7S9NEG6Y2xDtfZXTSZHu6E9YUPeeWApqU1x5BaX",
	"7Cahs398NJPkuckeaeChw4DJYCarwq1vMQ2qsYRZMNbn",
	"3kcMSx7ccb9vNW1Ftkb3ARxXCzhDSUHML38i4XGN3fKV",
	"CZr8UnwQrxXLz1BpVFQdDyCuRh75WDV4oTjqGRf4ziKj",
	"C44QSiUqrGUazCvQwcPMSutyKWo2Ev7xe1mhgXjFnRKT",
	"GD4JNXzzNKjVNaEh4hGK1LSD6hfFKTAtjQXXKi6qMR4F",
	"E89VNFjnah9SinP2yZtAnzwMjpomy52r77WuZRhWdyYS",
	"9KaVa9NA7UVG1w4djC1NbYusa22oGVN8BCG4jk8FB8Ma",
	"7EsNmZExFjJ2Ds4Yn6r9Xmpvd6aqu4Dte4V2tgT4pzzq",
	"5i4Y9DCaDxMtnR8XVZcYFvRM5dchj1pa8c5NH9RDBQbj",
	"3Mg679UwA2MDuAWoY5RT9Nj6YFYrJpS7YshYgLX5zidb",
	"DUaMhFXoALCF8PakY3wPPwKseb6EtsGMajypPoz6Qi8T",
	"3SuEmRCA2J2ARJEuVVsV96Ac1eLa5VMQBiPeQCQtziQo",
	"E2gZaCY6DzEvpqALRxtXeYXMiXGRQ4it9Pp4JcUGTws8",
	"2GRPKfzfm6kksRFdQQWTq5s8FfD8BNknGosFgfuV7ukX",
	"2KqvX96kQWUUpLmduPn5diiCRFkrG6CcrKxm78jNNvQK",
	"817dCUUsUQiVwsssQPa9JXA94cMY3kmnoJc6w3GwT2aX",
	"3tb7SDD8K6PXdh186TEQNZed95FUWZ5yCybxMeg1CdcK",
	"5PRyQP6KiUiHL3Zah5evgpYvzxZawHidEpx71MJ8VcvD",
	"EW8n6faxxYY5VyCWi1Y5Z9Di9zdBp8JeBJjwUtwuRD5e",
	"GQZ793sGJpgcPfT3Ss7FBNj46oRA5gSf4pmhemd1Rnvy",
	"9M5qhD7h3L2igs2e5cWxZjR8kgPqiDH7fJdxhTnky2Wi",
	"CbLTRB6RHuCoZyDyP1wzNoTqDqN4ZeTgxpks8gDMDxFP",
	"CLPcVTxseyUKGKVTBxiKHb7dJpitDeVpL8LDVD9aioLE",
	"2EKsJoYBdqk4WRNT7XTFiAAFDGNNhuNCdrM9xoAukQrL",
	"DNGRTGPh5oYBc3iR1GggXn4bnZTGZey2tDArR5gCQakB",
	"5MzHLA9XAbmd7GfxXBAb5KEsR5nnT2sLgSgcbebKpcC4",
	"39iPSB5EyysrzteCjpcxyyfWVV2sAJisAPRSvGJxmzCb",
	"6eHTVXpGpnRBypLw4z2oMgN8msYihcUM2KGTayM5Xq5r",
	"HMwn3E2KdNxSDLXHvZiRj9AA7FAWG8duAWWrbXcePEZV",
	"GBRN2BjaLXBEBGo5aAKiQPnCCgCK1xs7BKbC41M3ntQt",
	"oDdgkpDKAe2GJj916VFQ5gSc2SyTFKXh8dpUoCxjvPp",
	"GFyD74ZNrmUu5pfkJxtRSSknsTEiGqzu3tXKkBeddoWN",
	"3a1MkCVigbQBzNgw9v8RYf4jVTmzacxTPW3PJNV3yyVf",
	"B7qzrbb4wWcNL8wTBAbo6NDxutzAfmsYzx3eZ3Le5MaF",
	"EbgPEFQ9po7xjGDe4Lu5w3vMtFz2d96kTWwf3tZrkqs6",
	"4frZSPgzTCpwDuLpW1imwERNvEq5ajinmYf3BvT9HhYL",
	"AdHHkEsMG925QpLDdb7aPcA1NFkWMX8JP2p3SBPPiFjP",
	"DyLW9FprtDJT3DANSYz6PcWqXZxQgmT72U9fhJ5Hnrw4",
	"7zpzUCeMnThFEhDBWb8dqDsZ4674pLZpjAwquZnL2BJe",
	"Bbroz95o6KHJhzHLnQU8mjRhDMtGWvgoDXBeweNbqukY",
	"68vs5p3pBHLHBTBwHNPrwh6bVZH8B5iApihzpK3KgpJA",
	"EcqqFURMCbUY7DB4TGZXHhW1K3yBZbS3CNAPT2bN8Fux",
	"2tsngFMfURyQhiFYLVy9hvW2MH8esvUF7hQ1ktvybsRh",
	"97oxLmpzKFsA6UAU8eDmzGCB7mkyTvzphoX1exSBw8ub",
	"HjydJGzuHPeZ1CC2eNngws4LBfqGmHqRgWQaK9pPVt8N",
	"DESqpcJANfHaW6Uowbf5RNPGsqojB4vey6M46zQyHEc5",
	"83o8zyV56ZqUf32dTny5sWxBu81bK8Xdkgjz4Ui1H8dx",
	"G7tDWna5vQRJ4dz7dVhv62eKWJSGqrqufBKBdTht82oT",
	"CihP5vkT36FfEB3q9CCBQ338CYGZppsLMj99KpiVS6CW",
	"Cg1gtvfq7RJiCRJ4bfUVtZFVQmD5vJiRRwiT1PrNj6Ac",
	"BofXwcZLdCew4SkpEj4yg3pPD2z2NSQtgN5pdFppHoid",
	"DqbAMp8guRYhSWyLgLxqxw9LpaNukEQHArc86wzuvfWP",
	"aDmxKaS4FUDs8pepSXgRKjZ3tTmKez2FBKhnhWo3QMU",
	"G3BFQ9u6nZTS65bvDatLTxAsRBmYkNbjv6Q7jRqzTWcS",
	"9hcjKfYrsXAgBQdzQTXK4yQQNKUXoQCJTREmJspwx156",
	"72PGszpHNPrGs8VS7XtdZqGZcHUJCv6a8876JRXZJpjH",
	"5WnTr9zSCU1tAe9NRv67QFtg8mJYH9ZwsqWpZfrkNnNQ",
	"BHTDfrswif8D4VD8nP7WXJqY2wAWY6KE4YhVBAynPP1i",
	"2s8kcACv1RhdKtJXtf4YLSjNe8uokr9DoUYircK7Nw53",
	"3C5bN16KifQHitt2nTxURH6dAE1Ywd2WDbaAJdfoVHNX",
	"DhWnX4hHhUY38fkeBRBocP4uRQnM4AQFEyQjHqUWezhf",
	"9A9Ht9VTpu18KHaRMvQzxGMAUe8xtdvib66FVYar1kav",
	"GQdtbQt2MZKhff8CCd3dZKruK3aBe2wgH63kA2NYqUcG",
	"Cs4LU2v5cEdM9j2xpydx2GhmmoqPBZhkNUZqBPYS7Pje",
	"F8KoPmbbYGQTLSywZqrewBqVLAe9SJGBsWnyvFVDfWbh",
	"Aju5hbtXuWeLioxXrHHGNTJ5dmV5o2UtiubLpEjB7T8c",
	"DsDX3Tbzjrv4qsAFnrJuNDDVP9brjMXk6defzedKcyUn",
	"DfACCU2sG8Bo3F4GYKSeqPm9UAs5imK5428szfcXyxz",
	"CuTqpMNMfai8X2MSGVW35YLmXVu61wsh3CHEAoWsSMD3",
	"FdMvyzVxNxVhZxASqdPhbNeL7d4MVtrGFGdCaTcLmr6H",
	"5d3G5AtcbsKu1Jc85g1GtVmE8EBsoCuRbdX1L8xk8gZG",
	"HQUvZfSVmszkpop42H1mPAGWHGnRi1oxPhTY8n6cVTrK",
	"rto2X1SBvn8NxEZUrPX2tvze343EgcmnLDN4CkJwwQF",
	"CzYF1dz1M11r1Zw3c7qNdZW89irSznAjcbwpnQXDECCC",
	"HsjZSeo2HTMJjbenKRk3TBdxCG63XANiBbdd9aNLENEa",
	"5enN5s1PowwWcVq3XvrW48i6eNFnLH4BP2kSSHw93zFc",
	"FaZyJggC8mHQtLsf8c2oNhWkHjhzLKCW8m5ktLCZNdsH",
	"BfjXqv9xT9dD7wmp7HAokQGZhq9WFeihxewVFCm2gprR",
	"6RkGfHb1vmEtQygREMU9gpZqRaPDBeZ8735sza6cUZ8N",
	"CXjgC4BbfeiFxnZ7ibkbjTW1UiFRzoipudXcLJZAKrQM",
	"9pvqJHL19KNfp8nxjVkKAArcv3FLJfS7Tt9VAKj1scUL",
	"9ukU5qSBXxiyPZwv8nHrT8F8D3mPYyNZefZAEXdecVsJ",
	"H9AWRYwdBY8XZ938TNinbgY1ECMBYXRzEnYWbxiBcjzs",
	"7oBvy2zuXCZy5A3V8XeLtze3nEWRFaQYHDH4BSaDWNi1",
	"6NnqDu32ikx1Re5XyQRjtujcbHaUEQ9wBpN3Pg2jq2pT",
	"7PmTRcnaFqMHBnA255ZUP9o2nb5D3sT1ERUoVKhPMrUP",
	"79GmtFXgL25XFhYZM2M6Av3AuyLdQrhAFPRiXE8YPuuu",
	"7yUdJ3GEXqLGKBsnH4ashVob6MsjtcjTN9aXrWTzipK4",
	"GyUNA7XKycqhDuhJ7E5SSu5q83VuwALgHgaWDtrBz32k",
	"34EusgYvGWzs6zn7zpkVgZ7k531Ubx3CeBXDLsc9DCKX",
	"5tLmAVMfEDsFKhHhXKP5ALeDSUKdmNtu6bayQGWR9tMu",
	"8oFZZVxME1Tt9ZfgSZ8nbZB76qiVrMdmepiKF7C6n2uP",
	"BxEXEGBHZhNHZfEz3NTFXkb6gLf7KFDH1QjEsk5m221a",
	"4uuTDcGAX6hqFxp2B9VAdJnSSydBwF3NdMyyZrgArsGe",
	"EH9qmdvLvrnbD6uoTDxbT8Zwq2ZaPSYXtnAtqeLsXAXp",
	"FLhRBjMdk8BMLsEZHH7v5A9FciFarydJfmygzVoNgwC5",
	"EuTYz2vQcaPXGzmzbWjQb9fDgq34Sb3ZgLoU2sLJUnLg",
}

func TestMoreTXIDs(t *testing.T) {
	for _, in := range testIn {
		hash, _ := enc.ToBytes(in)
		id, err := types.ParseToTxID(hash)
		if err != nil {
			t.Errorf("Failed to parse TX hash %v : %v", enc.ToString(hash), err)
		} else if !bytes.Equal(hash, id[:]) {
			t.Errorf("TX hash %v and converted ID %v are differ.", enc.ToString(hash), id)
		}
	}
}
