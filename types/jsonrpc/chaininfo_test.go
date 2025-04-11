package jsonrpc

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/require"
)

func TestConvChainInfo(t *testing.T) {
	var (
		maxAer           = types.MaxAER
		stakingMinimum   = types.StakingMinimum
		stakingTotal     = types.NewAmount(10000000, types.Aergo)
		gasPrice         = types.NewAmount(5, types.Aergo)
		namePrice        = types.NewAmount(1, types.Aergo)
		totalVotingPower = types.NewAmount(10000000, types.Aergo)
		votingReward     = types.NewAmount(1, types.Aergo)
		hardfork         = map[string]uint64{"V1": 100000, "V2": 200000, "V3": 3000000}
	)

	for _, test := range []struct {
		types *types.ChainInfo
		inout *InOutChainInfo
	}{
		{&types.ChainInfo{ // not dpos - ignore stakingminimum, totalstaking
			BpNumber:         13,
			Maxblocksize:     uint64(chain.MaxBlockSize()),
			Maxtokens:        maxAer.Bytes(),
			Stakingminimum:   stakingMinimum.Bytes(),
			Totalstaking:     stakingTotal.Bytes(),
			Gasprice:         gasPrice.Bytes(),
			Nameprice:        namePrice.Bytes(),
			Totalvotingpower: totalVotingPower.Bytes(),
			Votingreward:     votingReward.Bytes(),
			Hardfork:         hardfork,
		}, &InOutChainInfo{
			BpNumber:         13,
			MaxBlockSize:     uint64(chain.MaxBlockSize()),
			MaxTokens:        maxAer.String(),
			GasPrice:         gasPrice.String(),
			NamePrice:        namePrice.String(),
			TotalVotingPower: totalVotingPower.String(),
			VotingReward:     votingReward.String(),
			Hardfork:         hardfork,
		}},

		{&types.ChainInfo{ // dpos
			Id:               &types.ChainId{Consensus: "dpos"},
			BpNumber:         13,
			Maxblocksize:     uint64(chain.MaxBlockSize()),
			Maxtokens:        maxAer.Bytes(),
			Stakingminimum:   stakingMinimum.Bytes(),
			Totalstaking:     stakingTotal.Bytes(),
			Gasprice:         gasPrice.Bytes(),
			Nameprice:        namePrice.Bytes(),
			Totalvotingpower: totalVotingPower.Bytes(),
			Votingreward:     votingReward.Bytes(),
			Hardfork:         hardfork,
		}, &InOutChainInfo{
			Id:               &InOutChainId{Consensus: "dpos"},
			BpNumber:         13,
			MaxBlockSize:     uint64(chain.MaxBlockSize()),
			MaxTokens:        maxAer.String(),
			StakingMinimum:   stakingMinimum.String(),
			Totalstaking:     stakingTotal.String(),
			GasPrice:         gasPrice.String(),
			NamePrice:        namePrice.String(),
			TotalVotingPower: totalVotingPower.String(),
			VotingReward:     votingReward.String(),
			Hardfork:         hardfork,
		}},
	} {
		require.Equal(t, test.inout, ConvChainInfo(test.types))
	}
}

func TestConvChainId(t *testing.T) {
	for _, test := range []struct {
		types *types.ChainId
		inout *InOutChainId
	}{
		{&types.ChainId{
			Version:   0,
			Magic:     "dev.chain",
			Public:    false,
			Mainnet:   false,
			Consensus: "sbp",
		}, &InOutChainId{
			Version:   0,
			Magic:     "dev.chain",
			Public:    false,
			Mainnet:   false,
			Consensus: "sbp",
		}},
	} {
		require.Equal(t, test.inout, ConvChainId(test.types))
	}
}

func TestConvBlockchainStatus(t *testing.T) {
	blockHash, err := base58.Decode(testBlockHashBase58)
	require.NoError(t, err)

	chainIdHash, err := base58.Decode(testChainIdHashBase58)
	require.NoError(t, err)

	consensusInfo := "consensus info"
	jsonConsensusInfo := json.RawMessage(consensusInfo)

	for _, test := range []struct {
		format string // base58 or hex
		types  *types.BlockchainStatus
		inout  *InOutBlockchainStatus
	}{
		{"base58", &types.BlockchainStatus{
			BestBlockHash:   blockHash,
			BestHeight:      1,
			ConsensusInfo:   consensusInfo,
			BestChainIdHash: chainIdHash,
		}, &InOutBlockchainStatus{
			Hash:          testBlockHashBase58,
			Height:        1,
			ChainIdHash:   testChainIdHashBase58,
			ConsensusInfo: &jsonConsensusInfo,
		}},
		{"hex", &types.BlockchainStatus{
			BestBlockHash:   blockHash,
			BestHeight:      1,
			ConsensusInfo:   consensusInfo,
			BestChainIdHash: chainIdHash,
		}, &InOutBlockchainStatus{
			Hash:        hex.EncodeToString(blockHash),
			Height:      1,
			ChainIdHash: hex.EncodeToString(chainIdHash),
		}},
	} {
		if test.format == "base58" {
			require.Equal(t, test.inout, ConvBlockchainStatus(test.types))
		} else if test.format == "hex" {
			require.Equal(t, test.inout, ConvHexBlockchainStatus(test.types))
		} else {
			require.Fail(t, "invalid format", test.format)
		}
	}
}
