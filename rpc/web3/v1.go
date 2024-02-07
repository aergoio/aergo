package web3

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aergoio/aergo/v2/contract/system"
	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/rpc"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/jsonrpc"
)

type APIHandler func() (http.Handler, bool)

type Web3APIv1 struct {
	rpc     *rpc.AergoRPCService
	request *http.Request

	handlerMap map[string]map[string]APIHandler
}

func (api *Web3APIv1) NewHandler() {
	handlerGet := map[string]APIHandler{
		"/getState":                api.GetState,
		"/getStateAndProof":        api.GetStateAndProof,
		"/getBalance":              api.GetBalance,
		"/getBlock":                api.GetBlock,
		"/blockchain":              api.Blockchain,
		"/getBlockTx":              api.GetBlockBody,
		"/listBlockHeaders":        api.ListBlockHeaders,
		"/getBlockMetadata":        api.GetBlockMetadata,
		"/getTx":                   api.GetTX,
		"/getReceipt":              api.GetReceipt,
		"/getReceipts":             api.GetReceipts,
		"/queryContract":           api.QueryContract,
		"/listEvents":              api.ListEvents,
		"/getABI":                  api.GetABI,
		"/queryContractStateProof": api.QueryContractState,
		"/getTxCount":              api.GetBlockTransactionCount,
		"/getChainInfo":            api.GetChainInfo,
		"/getConsensusInfo":        api.GetConsensusInfo,
		"/getNodeInfo":             api.NodeState,
		"/getChainId":              api.GetChainId,
		"/getPeers":                api.GetPeers,
		"/getServerInfo":           api.GetServerInfo,
		"/metric":                  api.Metric,
		"/chainStat":               api.ChainStat,
	}

	ca := api.rpc.GetActorHelper().GetChainAccessor()
	consensus := &jsonrpc.InOutConsensusInfo{}
	json.Unmarshal([]byte(ca.GetConsensusInfo()), consensus)

	if consensus.Type == "raft" {
		handlerGet["/getEnterpriseConfig"] = api.GetEnterpriseConfig
	} else if consensus.Type == "dpos" {
		handlerGet["/getStaking"] = api.GetStaking
		handlerGet["/getVotes"] = api.GetVotes
		handlerGet["/getAccountVotes"] = api.GetAccountVotes
		handlerGet["/getNameInfo"] = api.GetNameInfo
	}

	handlerPost := map[string]APIHandler{
		"/sendSignedTransaction": api.CommitTX,
	}

	api.handlerMap = make(map[string]map[string]APIHandler)
	api.handlerMap[http.MethodGet] = handlerGet
	api.handlerMap[http.MethodPost] = handlerPost
}

func (api *Web3APIv1) handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rc := recover(); rc != nil {
			logger.Error().Msg("panic web3 : " + r.URL.Path + "?" + r.URL.RawQuery)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	api.request = r
	handler, ok := api.restAPIHandler(r)
	if ok {
		handler.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (api *Web3APIv1) restAPIHandler(r *http.Request) (handler http.Handler, ok bool) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, prefixV1)

	selectedHandler := api.handlerMap[r.Method][path]

	if selectedHandler != nil {

		return selectedHandler()
	}

	return nil, false
}

func (api *Web3APIv1) GetState() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.SingleBytes{}
	account := values.Get("account")
	if account != "" {
		accountBytes, err := types.DecodeAddress(account)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = accountBytes
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: account"), http.StatusBadRequest), true
	}

	result, err := api.rpc.GetState(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	output := jsonrpc.ConvState(result)
	output.Account = account

	// Staking, NextAction
	resultStaking, err := api.rpc.GetStaking(api.request.Context(), &types.AccountAddress{Value: request.Value})
	if err == nil {
		output.Stake = new(big.Int).SetBytes(resultStaking.GetAmount()).String()
		output.When = resultStaking.GetWhen()
		if output.When != 0 {
			output.NextAction = output.When + 86400
		}
	}

	var balanceInt, stakeInt *big.Int
	if output.Balance != "" {
		balanceInt, _ = new(big.Int).SetString(output.Balance, 10)
	} else {
		balanceInt = new(big.Int)
	}

	if output.Stake != "" {
		stakeInt, _ = new(big.Int).SetString(output.Stake, 10)
	} else {
		stakeInt = new(big.Int)
	}

	totalInt := new(big.Int).Add(stakeInt, balanceInt)
	output.Total = totalInt.String()

	// VotingPower
	resultVotingPower, err := api.rpc.GetAccountVotes(api.request.Context(), &types.AccountAddress{Value: request.Value})
	if err == nil && resultVotingPower.Voting != nil {
		sum := new(big.Int)
		for _, vote := range resultVotingPower.Voting {
			m, ok := new(big.Int).SetString(vote.GetAmount(), 10)
			if ok {
				sum.Add(sum, m)
			}
		}
		output.VotingPower = sum.String()
	}

	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetStateAndProof() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.AccountAndRoot{}
	account := values.Get("account")
	if account != "" {
		accountBytes, err := types.DecodeAddress(account)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Account = accountBytes
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: account"), http.StatusBadRequest), true
	}

	compressed := values.Get("compressed")
	if compressed != "" {
		compressedValue, parseErr := strconv.ParseBool(compressed)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Compressed = compressedValue
	}

	result, err := api.rpc.GetStateAndProof(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvStateAndPoof(result)
	output.Account = account
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetNameInfo() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.Name{}
	name := values.Get("name")
	if name != "" {
		request.Name = name

		if len(name) > types.NameLength {
			return commonResponseHandler(&types.Empty{}, errors.New("name is not in format 12 digits")), true
		}
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: name"), http.StatusBadRequest), true
	}

	number := values.Get("number")
	if number != "" {
		numberValue, parseErr := strconv.ParseUint(number, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.BlockNo = numberValue
	}

	result, err := api.rpc.GetNameInfo(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvNameInfo(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetBalance() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.SingleBytes{}
	account := values.Get("account")
	if account != "" {
		accountBytes, err := types.DecodeAddress(account)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = accountBytes
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: account"), http.StatusBadRequest), true
	}

	result, err := api.rpc.GetState(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvBalance(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetBlock() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.SingleBytes{}
	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base58.Decode(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	}

	number := values.Get("number")
	if number != "" {
		numberValue, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		number := uint64(numberValue) // Replace with your actual value
		byteValue := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteValue, number)
		request.Value = byteValue
	}

	result, err := api.rpc.GetBlock(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvBlock(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetBlockTransactionCount() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.SingleBytes{}
	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base58.Decode(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	}

	number := values.Get("number")
	if number != "" {
		numberValue, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		number := uint64(numberValue) // Replace with your actual value
		byteValue := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteValue, number)
		request.Value = byteValue
	}

	result, err := api.rpc.GetBlock(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvBlockTransactionCount(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) Blockchain() (handler http.Handler, ok bool) {
	ca := api.rpc.GetActorHelper().GetChainAccessor()
	last, err := ca.GetBestBlock()
	if err != nil {
		return nil, false
	}

	digest := sha256.New()
	digest.Write(last.GetHeader().GetChainID())
	bestChainIDHash := digest.Sum(nil)

	chainInfo, err := api.rpc.GetChainInfo(api.request.Context(), &types.Empty{})
	if err != nil {
		logger.Warn().Err(err).Msg("failed to get chain info in blockchain")
		chainInfo = nil
	}

	result := &types.BlockchainStatus{
		BestBlockHash:   last.BlockHash(),
		BestHeight:      last.GetHeader().GetBlockNo(),
		ConsensusInfo:   ca.GetConsensusInfo(),
		BestChainIdHash: bestChainIDHash,
		ChainInfo:       chainInfo,
	}

	output := jsonrpc.ConvBlockchainStatus(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetBlockBody() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.BlockBodyParams{}
	request.Paging = &types.PageParams{}

	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base58.Decode(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Hashornumber = hashBytes
	}

	number := values.Get("number")
	if number != "" {
		numberValue, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		number := uint64(numberValue)
		byteValue := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteValue, number)
		request.Hashornumber = byteValue
	}

	size := values.Get("size")
	if size != "" {
		sizeValue, parseErr := strconv.ParseUint(size, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true

		}
		request.Paging.Size = uint32(sizeValue)
		if request.Paging.Size > 100 {
			request.Paging.Size = 100
		}
	}

	offset := values.Get("offset")
	if offset != "" {
		offsetValue, parseErr := strconv.ParseUint(offset, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Paging.Offset = uint32(offsetValue)
	}

	result, err := api.rpc.GetBlockBody(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvBlockBodyPaged(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) ListBlockHeaders() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.ListParams{}
	height := values.Get("height")
	if height != "" {
		heightValue, parseErr := strconv.ParseUint(height, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Height = heightValue
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: height"), http.StatusBadRequest), true
	}

	size := values.Get("size")
	if size != "" {
		sizeValue, parseErr := strconv.ParseUint(size, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Size = uint32(sizeValue)
		if request.Size > 100 {
			request.Size = 100
		}
	} else {
		request.Size = 1
	}

	offset := values.Get("offset")
	if offset != "" {
		offsetValue, parseErr := strconv.ParseUint(offset, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Offset = uint32(offsetValue)
	}

	asc := values.Get("asc")
	if asc != "" {
		ascValue, parseErr := strconv.ParseBool(asc)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Asc = ascValue
	}

	result, err := api.rpc.ListBlockHeaders(api.request.Context(), request)
	if err != nil {
		errStr := err.Error()

		strBlockNo := strings.TrimPrefix(errStr, "block not found: blockNo=")
		if strBlockNo == errStr {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		return commonResponseHandler(&types.Empty{}, errors.New("rpc error: code = Internal desc = "+errStr)), true
	}

	output := jsonrpc.ConvBlockHeaderList(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetBlockMetadata() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.SingleBytes{}
	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base58.Decode(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	}

	number := values.Get("number")
	if number != "" {
		numberValue, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		number := uint64(numberValue) // Replace with your actual value
		byteValue := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteValue, number)
		request.Value = byteValue
	}

	result, err := api.rpc.GetBlockMetadata(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvBlockMetadata(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) ListBlockMetadata() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.ListParams{}
	height := values.Get("height")
	if height != "" {
		heightValue, parseErr := strconv.ParseUint(height, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Height = heightValue
	}

	size := values.Get("size")
	if size != "" {
		sizeValue, parseErr := strconv.ParseUint(size, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Size = uint32(sizeValue)
		if request.Size > 100 {
			request.Size = 100
		}
	}

	offset := values.Get("offset")
	if offset != "" {
		offsetValue, parseErr := strconv.ParseUint(offset, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Offset = uint32(offsetValue)
	}

	asc := values.Get("asc")
	if asc != "" {
		ascValue, parseErr := strconv.ParseBool(asc)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Asc = ascValue
	}

	result, err := api.rpc.ListBlockMetadata(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvListBlockMetadata(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetChainInfo() (handler http.Handler, ok bool) {
	request := &types.Empty{}

	result, err := api.rpc.GetChainInfo(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvChainInfo(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) ChainStat() (handler http.Handler, ok bool) {
	request := &types.Empty{}

	result, err := api.rpc.ChainStat(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvChainStat(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetConsensusInfo() (handler http.Handler, ok bool) {
	request := &types.Empty{}

	result, err := api.rpc.GetConsensusInfo(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvConsensusInfo(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetReceipt() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.SingleBytes{}
	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base58.Decode(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: hash"), http.StatusBadRequest), true
	}

	result, err := api.rpc.GetReceipt(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.MarshalJSON(result), nil), true
}

func (api *Web3APIv1) GetReceipts() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.ReceiptsParams{}
	request.Paging = &types.PageParams{}

	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base58.Decode(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Hashornumber = hashBytes
	}

	number := values.Get("number")
	if number != "" {
		numberValue, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		number := uint64(numberValue) // Replace with your actual value
		byteValue := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteValue, number)
		request.Hashornumber = byteValue
	}

	size := values.Get("size")
	if size != "" {
		sizeValue, parseErr := strconv.ParseUint(size, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true

		}
		request.Paging.Size = uint32(sizeValue)
		if request.Paging.Size > 100 {
			request.Paging.Size = 100
		}
	}

	offset := values.Get("offset")
	if offset != "" {
		offsetValue, parseErr := strconv.ParseUint(offset, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Paging.Offset = uint32(offsetValue)
	}

	result, err := api.rpc.GetReceipts(api.request.Context(), request)
	if err != nil {
		errStr := err.Error()

		strBlockNo := strings.TrimPrefix(errStr, "empty : blockNo=")
		if strBlockNo == errStr {
			return commonResponseHandler(&types.Empty{}, err), true
		}

		blockNo, err := strconv.ParseUint(strBlockNo, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}

		receipts := &jsonrpc.InOutReceipts{}
		receipts.BlockNo = blockNo
		return stringResponseHandler(jsonrpc.MarshalJSON(receipts), nil), true
	}

	output := jsonrpc.ConvReceiptsPaged(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetTX() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.SingleBytes{}
	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base58.Decode(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: hash"), http.StatusBadRequest), true
	}

	result, err := api.rpc.GetTX(api.request.Context(), request)
	if err == nil {
		output := jsonrpc.ConvTx(result, jsonrpc.Base58)
		jsonrpc.CovPayloadJson(output)
		return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
	} else {
		resultblock, err := api.rpc.GetBlockTX(api.request.Context(), request)

		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}

		output := jsonrpc.ConvTxInBlock(resultblock, jsonrpc.Base58)
		jsonrpc.CovPayloadJson(output.Tx)
		return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
	}
}

func (api *Web3APIv1) QueryContract() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.Query{}
	address := values.Get("address")
	if address != "" {
		hashBytes, err := types.DecodeAddress(address)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.ContractAddress = hashBytes
	}

	var ci types.CallInfo
	name := values.Get("name")
	if name != "" {
		ci.Name = name
	}

	query := values.Get("query")

	if query != "" {
		err = json.Unmarshal([]byte(query), &ci.Args)

		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
	}

	callinfo, err := json.Marshal(ci)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request.Queryinfo = callinfo

	result, err := api.rpc.QueryContract(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvQueryContract(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) ListEvents() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.FilterInfo{}
	address := values.Get("address")
	if address != "" {
		hashBytes, err := types.DecodeAddress(address)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.ContractAddress = hashBytes
	}

	EventName := values.Get("eventName")
	if EventName != "" {
		request.EventName = EventName
	}

	Blockfrom := values.Get("blockfrom")
	if Blockfrom != "" {
		BlockfromValue, err := strconv.ParseUint(Blockfrom, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Blockfrom = uint64(BlockfromValue)
	}

	Blockto := values.Get("blockto")
	if Blockto != "" {
		BlocktoValue, err := strconv.ParseUint(Blockto, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Blockto = uint64(BlocktoValue)
	}

	desc := values.Get("desc")
	if desc != "" {
		descValue, parseErr := strconv.ParseBool(desc)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Desc = descValue
	}

	argFilter := values.Get("argFilter")
	if argFilter != "" {
		request.ArgFilter = []byte(argFilter)
	}

	recentBlockCnt := values.Get("recentBlockCnt")
	if recentBlockCnt != "" {
		recentBlockCntValue, parseErr := strconv.ParseInt(recentBlockCnt, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.RecentBlockCnt = int32(recentBlockCntValue)
		if request.RecentBlockCnt > 10000 {
			request.RecentBlockCnt = 10000
		}
	} else if Blockfrom == "" && Blockto == "" {
		request.RecentBlockCnt = 10000
	}

	result, err := api.rpc.ListEvents(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	output := jsonrpc.ConvEvents(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetABI() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.SingleBytes{}
	address := values.Get("address")
	if address != "" {
		hashBytes, err := types.DecodeAddress(address)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	}

	result, err := api.rpc.GetABI(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvAbi(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetAccountVotes() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.AccountAddress{}
	account := values.Get("account")
	if account != "" {
		hashBytes, err := types.DecodeAddress(account)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: account"), http.StatusBadRequest), true
	}

	result, err := api.rpc.GetAccountVotes(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvInOutAccountVoteInfo(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) QueryContractState() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.StateQuery{}
	address := values.Get("address")
	if address != "" {
		addressBytes, err := types.DecodeAddress(address)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.ContractAddress = addressBytes
	} else {
		return commonResponseHandlerWithCode(&types.Empty{}, errors.New("Missing required parameter: address"), http.StatusBadRequest), true
	}

	storageKeyPlain := bytes.NewBufferString("_sv_")
	args1 := values.Get("varname1")
	if args1 != "" {
		storageKeyPlain.WriteString(args1)
	}
	args2 := values.Get("varname2")
	if args2 != "" {
		storageKeyPlain.WriteString("-")
		storageKeyPlain.WriteString(args2)
	}

	storageKey := common.Hasher([]byte(storageKeyPlain.Bytes()))
	request.StorageKeys = [][]byte{storageKey}

	compressed := values.Get("compressed")
	if compressed != "" {
		compressedValue, parseErr := strconv.ParseBool(compressed)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Compressed = compressedValue
	}

	result, err := api.rpc.QueryContractState(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvQueryContractState(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) NodeState() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.NodeReq{}
	component := values.Get("component")
	if component != "" {
		request.Component = []byte(component)
	}

	timeout := values.Get("timeout")
	byteValue := make([]byte, 8)
	if timeout != "" {
		timeoutValue, err := strconv.ParseUint(timeout, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		timeout := uint64(timeoutValue)
		binary.LittleEndian.PutUint64(byteValue, timeout)
	} else {
		binary.LittleEndian.PutUint64(byteValue, 3)
	}
	request.Timeout = byteValue

	result, err := api.rpc.NodeState(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvNodeState(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetPeers() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.PeersParams{}
	noHidden := values.Get("noHidden")
	if noHidden != "" {
		noHiddenValue, parseErr := strconv.ParseBool(noHidden)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.NoHidden = noHiddenValue
	}

	showSelf := values.Get("showSelf")
	if showSelf != "" {
		showSelfValue, parseErr := strconv.ParseBool(showSelf)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.ShowSelf = showSelfValue
	}

	result, err := api.rpc.GetPeers(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvPeerList(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetChainId() (handler http.Handler, ok bool) {
	chainInfo, err := api.rpc.GetChainInfo(api.request.Context(), &types.Empty{})
	if err != nil {
		logger.Warn().Err(err).Msg("failed to get chain info in blockchain")
		chainInfo = nil
	}

	output := jsonrpc.ConvChainInfo(chainInfo)
	return stringResponseHandler(jsonrpc.MarshalJSON(output.Id), nil), true
}

func (api *Web3APIv1) GetServerInfo() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.KeyParams{}
	keys := values["key"]
	if len(keys) > 0 {
		request.Key = keys
	}

	result, err := api.rpc.GetServerInfo(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvServerInfo(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetStaking() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.AccountAddress{}
	address := values.Get("address")
	if address != "" {
		addressBytes, err := types.DecodeAddress(address)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = addressBytes
	}

	result, err := api.rpc.GetStaking(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvStaking(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) GetVotes() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	getCount := func(id string, count uint32) uint32 {
		if strings.ToLower(id) == strings.ToLower(types.OpvoteBP.ID()) {
			return count
		}

		if count == 0 {
			return 1
		}
		return count
	}

	request := &types.VoteParams{}
	request.Id = types.OpvoteBP.ID()

	id := values.Get("id")
	if id != "" {
		request.Id = id
	}

	reqCount := uint32(0)
	count := values.Get("count")
	if count != "" {
		sizeValue, parseErr := strconv.ParseUint(count, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		reqCount = uint32(sizeValue)
	}

	var output []*jsonrpc.InOutVotes
	if id == "" {
		for _, i := range system.GetVotingCatalog() {
			id := i.ID()

			if id != "" {
				result, err := api.rpc.GetVotes(api.request.Context(), &types.VoteParams{
					Id:    id,
					Count: getCount(id, reqCount),
				})

				if err == nil {
					subOutput := jsonrpc.ConvVotes(result, id)
					subOutput.Id = id
					output = append(output, subOutput)
				}
			}
		}
	} else {
		request.Count = getCount(id, reqCount)
		result, err := api.rpc.GetVotes(api.request.Context(), request)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}

		subOutput := jsonrpc.ConvVotes(result, id)
		subOutput.Id = id
		output = append(output, subOutput)
	}

	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) Metric() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.MetricsRequest{}
	metricType := values.Get("type")
	if metricType != "" {
		request.Types = append(request.Types, types.MetricType(types.MetricType_value[metricType]))
	} else {
		request.Types = append(request.Types, types.MetricType(types.MetricType_value["P2P_NETWORK"]))
	}

	result, err := api.rpc.Metric(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	res := jsonrpc.ConvMetrics(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(res), nil), true
}

func (api *Web3APIv1) GetEnterpriseConfig() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.EnterpriseConfigKey{}

	key := values.Get("key")
	if key != "" {
		request.Key = key
	}

	result, err := api.rpc.GetEnterpriseConfig(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvEnterpriseConfig(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}

func (api *Web3APIv1) CommitTX() (handler http.Handler, ok bool) {
	body, err := io.ReadAll(api.request.Body)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	txs, err := jsonrpc.ParseBase58Tx([]byte(body))
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	if len(txs) > 0 && txs[0].Body.Type == types.TxType_DEPLOY {
		return commonResponseHandler(&types.Empty{}, errors.New("Feature not supported yet")), true
	}

	result, err := api.rpc.CommitTX(api.request.Context(), &types.TxList{Txs: txs})
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	output := jsonrpc.ConvCommitResultList(result)
	return stringResponseHandler(jsonrpc.MarshalJSON(output), nil), true
}
