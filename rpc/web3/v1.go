package web3

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aergoio/aergo/v2/internal/common"
	"github.com/aergoio/aergo/v2/rpc"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/jsonrpc"
	"github.com/asaskevich/govalidator"
	"github.com/mr-tron/base58"
)

type Web3APIv1 struct {
	rpc     *rpc.AergoRPCService
	request *http.Request
}

func (api *Web3APIv1) handler(w http.ResponseWriter, r *http.Request) {
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

	if r.Method == http.MethodGet {
		switch path {
		case "/getAccounts":
			return api.GetAccounts()
		case "/getState":
			return api.GetState()
		case "/getProof":
			return api.GetStateAndProof()
		case "/getNameInfo":
			return api.GetNameInfo()
		case "/getBalance":
			return api.GetBalance()
		case "/getBlock":
			return api.GetBlock()
		case "/getBlockNumber":
			return api.Blockchain()
		case "/getBlockBody":
			return api.GetBlockBody()
		case "/listBlockHeaders":
			return api.ListBlockHeaders()
		case "/getBlockMetadata":
			return api.GetBlockMetadata()
		case "/getTransaction":
			return api.GetTX()
		case "/getTransactionReceipt":
			return api.GetReceipt()
		case "/getBlockTransactionReceipts":
			return api.GetReceipts()
		case "/getBlockTX":
			return api.GetBlockTX()
		case "/call":
			return api.QueryContract()
		case "/getPastEvents":
			return api.ListEvents()
		case "/getABI":
			return api.GetABI()
		case "/queryContractState":
			return api.QueryContractState()
		case "/getBlockTransactionCount":
			return api.GetBlockTransactionCount()
		case "/getChainInfo":
			return api.GetChainInfo()
		case "/getConsensusInfo":
			return api.GetConsensusInfo()
		case "/getAccountVotes":
			return api.GetAccountVotes()
		case "/getNodeInfo":
			return api.NodeState()
		case "/getChainId":
			return api.GetPeers()
		case "/getServerInfo":
			return api.GetServerInfo()
		case "/getStaking":
			return api.GetStaking()
		case "/getVotes":
			return api.GetVotes()
		case "/metric":
			return api.Metric()
		case "/getEnterpriseConfig":
			return api.GetEnterpriseConfig()
		case "/getConfChangeProgress":
			return api.GetConfChangeProgress()
		case "/chainStat":
			return api.ChainStat()
		default:
			return nil, false
		}
	} else if r.Method == http.MethodPost {
		switch path {
		case "/sendSignedTransaction":
			return api.CommitTX()
		default:
			return nil, false
		}
	}
	return nil, false
}

func (api *Web3APIv1) GetAccounts() (handler http.Handler, ok bool) {
	request := &types.Empty{}

	msg, err := api.rpc.GetAccounts(api.request.Context(), request)

	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	out := fmt.Sprintf("%s", "[")
	if msg != nil {
		addresslist := msg.GetAccounts()
		for _, a := range addresslist {
			out = fmt.Sprintf("%s\"%s\", ", out, types.EncodeAddress(a.Address))
		}
		if addresslist != nil && len(out) >= 2 {
			out = out[:len(out)-2]
		}
	}
	out = fmt.Sprintf("%s%s", out, "]")

	return stringResponseHandler(out, nil), true
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
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetState(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	balance, err := jsonrpc.ConvertUnit(msg.GetBalanceBigInt(), "aergo")
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result := fmt.Sprintf(`{"account":"%s", "balance":"%s", "nonce":%d}`, account, balance, msg.GetNonce())
	return stringResponseHandler(result, nil), true
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
	}

	compressed := values.Get("compressed")
	if compressed != "" {
		compressedValue, parseErr := strconv.ParseBool(compressed)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Compressed = compressedValue
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetStateAndProof(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	balance, err := jsonrpc.ConvertUnit(msg.GetState().GetBalanceBigInt(), "aergo")
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result := fmt.Sprintf(`{"account":"%s", "nonce":%d, "balance":"%s", "included":%t, "merkle proof length":%d, "height":%d}`+"\n",
		account, msg.GetState().GetNonce(), balance, msg.GetInclusion(), len(msg.GetAuditPath()), msg.GetHeight())
	return stringResponseHandler(result, nil), true
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
	}
	number := values.Get("number")
	if number != "" {
		numberValue, parseErr := strconv.ParseUint(number, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.BlockNo = numberValue
	}

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetNameInfo(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result := fmt.Sprintf(`{"%s": {"Owner" : "%s", "Destination" : "%s" }}`,
		msg.Name.Name, types.EncodeAddress(msg.Owner), types.EncodeAddress(msg.Destination))

	return stringResponseHandler(result, nil), true
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
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetState(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result := fmt.Sprintf(`{"balance":"%s"}`, msg.GetBalanceBigInt())
	return stringResponseHandler(result, nil), true
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

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.GetBlock(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetBlock(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result := fmt.Sprintf(`{"count":"%d"}`, func() int {
		if msg.Body.Txs != nil {
			return len(msg.Body.Txs)
		}
		return 0
	}())

	return stringResponseHandler(result, nil), true
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

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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
	}

	offset := values.Get("offset")
	if offset != "" {
		offsetValue, parseErr := strconv.ParseUint(offset, 10, 64)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Paging.Offset = uint32(offsetValue)
	}

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.GetBlockBody(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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
	}

	size := values.Get("size")
	if size != "" {
		sizeValue, parseErr := strconv.ParseUint(size, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Size = uint32(sizeValue)
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

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.ListBlockHeaders(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.GetBlockMetadata(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	return commonResponseHandler(api.rpc.ListBlockMetadata(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetChainInfo() (handler http.Handler, ok bool) {
	request := &types.Empty{}
	return commonResponseHandler(api.rpc.GetChainInfo(api.request.Context(), request)), true
}

func (api *Web3APIv1) ChainStat() (handler http.Handler, ok bool) {
	request := &types.Empty{}
	return commonResponseHandler(api.rpc.ChainStat(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetConsensusInfo() (handler http.Handler, ok bool) {
	request := &types.Empty{}
	return commonResponseHandler(api.rpc.GetConsensusInfo(api.request.Context(), request)), true
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
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	return commonResponseHandler(api.rpc.GetReceipt(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetReceipts() (handler http.Handler, ok bool) {
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	return commonResponseHandler(api.rpc.GetReceipts(api.request.Context(), request)), true
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
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetTX(api.request.Context(), request)
	if err == nil {
		return commonResponseHandler(jsonrpc.ConvTxEx(msg, jsonrpc.Base58), nil), true
	} else {
		msgblock, err := api.rpc.GetBlockTX(api.request.Context(), request)

		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		return commonResponseHandler(jsonrpc.ConvTxInBlockEx(msgblock, jsonrpc.Base58), nil), true
	}

}

func (api *Web3APIv1) GetBlockTX() (handler http.Handler, ok bool) {
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
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetTX(api.request.Context(), request)
	if err == nil {
		return commonResponseHandler(jsonrpc.ConvTxEx(msg, jsonrpc.Base58), nil), true

	} else {
		msgblock, err := api.rpc.GetBlockTX(api.request.Context(), request)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		return commonResponseHandler(jsonrpc.ConvTxInBlockEx(msgblock, jsonrpc.Base58), nil), true
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.QueryContract(api.request.Context(), request)
	return stringResponseHandler("{\"result\":"+string(msg.Value)+"}", nil), true
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

	EventName := values.Get("EventName")
	if EventName != "" {
		request.EventName = EventName
	}

	Blockfrom := values.Get("Blockfrom")
	if Blockfrom != "" {
		BlockfromValue, err := strconv.ParseUint(Blockfrom, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Blockfrom = uint64(BlockfromValue)
	}

	Blockto := values.Get("Blockto")
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
	} else {
		request.RecentBlockCnt = 0
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return commonResponseHandler(api.rpc.ListEvents(api.request.Context(), request)), true
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

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return commonResponseHandler(api.rpc.GetABI(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetAccountVotes() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.AccountAddress{}
	address := values.Get("address")
	if address != "" {
		hashBytes, err := types.DecodeAddress(address)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	}

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return commonResponseHandler(api.rpc.GetAccountVotes(api.request.Context(), request)), true
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.QueryContractState(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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
	if timeout != "" {
		timeoutValue, err := strconv.ParseUint(timeout, 10, 64)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		timeout := uint64(timeoutValue) // Replace with your actual value
		byteValue := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteValue, timeout)
		request.Timeout = byteValue
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.NodeState(api.request.Context(), request)

	return stringResponseHandler(string(msg.Value), nil), true
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.GetPeers(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.GetServerInfo(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return commonResponseHandler(api.rpc.GetStaking(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetVotes() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.VoteParams{}
	request.Id = types.OpvoteBP.ID()

	count := values.Get("count")
	if count != "" {
		sizeValue, parseErr := strconv.ParseUint(count, 10, 32)
		if parseErr != nil {
			return commonResponseHandler(&types.Empty{}, parseErr), true
		}
		request.Count = uint32(sizeValue)
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetVotes(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result := "["
	comma := ","
	for i, r := range msg.GetVotes() {
		result = result + "{\"" + base58.Encode(r.Candidate) + "\":" + r.GetAmountBigInt().String() + "}"
		if i+1 == len(msg.GetVotes()) {
			comma = ""
		}
		result = result + comma
	}
	result = result + "]"

	return stringResponseHandler(result, nil), true
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
	}

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	result, err := api.rpc.Metric(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return stringResponseHandler(jsonrpc.JSON(result), nil), true
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

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	msg, err := api.rpc.GetEnterpriseConfig(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	type outConf struct {
		Key    string
		On     *bool
		Values []string
	}

	var out outConf
	out.Key = msg.Key
	out.Values = msg.Values
	if strings.ToUpper(key) != "PERMISSIONS" {
		out.On = &msg.On
	}

	return stringResponseHandler(jsonrpc.B58JSON(out), nil), true
}

func (api *Web3APIv1) GetConfChangeProgress() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	request := &types.SingleBytes{}
	hash := values.Get("hash")
	if hash != "" {
		hashBytes, err := base64.StdEncoding.DecodeString(hash)
		if err != nil {
			return commonResponseHandler(&types.Empty{}, err), true
		}
		request.Value = hashBytes
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	block, err := api.rpc.GetBlock(api.request.Context(), request)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, block.BlockNo())

	return commonResponseHandler(api.rpc.GetConfChangeProgress(api.request.Context(), &types.SingleBytes{Value: b})), true
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

	return commonResponseHandler(api.rpc.CommitTX(api.request.Context(), &types.TxList{Txs: txs})), true
}
