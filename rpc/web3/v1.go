package web3

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aergoio/aergo/rpc"
	"github.com/aergoio/aergo/types"
	"github.com/asaskevich/govalidator"
	"github.com/mr-tron/base58"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Web3APIv1 struct {
	rpc *rpc.AergoRPCService
	request *http.Request
}

func (api *Web3APIv1) handler(w http.ResponseWriter, r *http.Request) {
    api.request = r;

	handler, ok := api.restAPIHandler(r)	
	if(ok) {
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
			case "/getAccounts":				return api.GetAccounts();
			case "/getState":					return api.GetState();
			case "/getProof":					return api.GetStateAndProof();
			case "/getNameInfo":				return api.GetNameInfo();
			case "/getBlock":					return api.GetBlock();
			case "/getBlockNumber":				return api.Blockchain();
			case "/getBlockBody":				return api.GetBlockBody();
			case "/listBlockHeaders":			return api.ListBlockHeaders();
			case "/getBlockMetadata":			return api.GetBlockMetadata();
			case "/getChainInfo":				return api.GetChainInfo();	
			case "/getConsencusInfo":			return api.GetConsensusInfo();
			case "/getTransaction":				return api.GetTX();
			case "/getTransactionReceipt":		return api.GetReceipt();
			case "/getBlockTX":					return api.GetBlockTX();
			case "/verifyTX":					return api.VerifyTX();
			case "/call":						return api.QueryContract();
			case "/getPastEvents":				return api.ListEvents();
			case "/getABI":						return api.GetABI();
			case "/getAccountVotes":			return api.GetAccountVotes();
			case "/queryContractState":			return api.QueryContractState();
			case "/getNodeInfo":				return api.NodeState();
			case "/getChainId":					return api.GetPeers();
			case "/getServerInfo":				return api.GetServerInfo();
			case "/getStaking":					return api.GetStaking();
			case "/getVotes":					return api.GetVotes();
			case "/metric":						return api.Metric();
			case "/getEnterpriseConfig":		return api.GetEnterpriseConfig();
			case "/getConfChangeProgress":		return api.GetConfChangeProgress();

			// case "/getBalance":						return api.GetBalance();
			// case "/getBlockTransactionCount":		return api.GetBlockTransactionCount();
			// case "/getTransactionCount":			return api.GetTransactionCount();
			// case "/getBlockTransactionReceipts":	return api.GetBlockTransactionReceipts();
			
			case "/ChainStat":					return api.ChainStat();
			case "/ListBlockMetadata":			return api.ListBlockMetadata();
			
			default:							return nil, false
		}
	} else if r.Method == http.MethodPost {
		switch path {
			case "/sendSignedTransaction":		return api.CommitTX();
			default:							return nil, false
		}
	}
	return nil, false
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

	return commonResponseHandler(api.rpc.Metric(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetAccounts() (handler http.Handler, ok bool) {
	request := &types.Empty{}
	return commonResponseHandler(api.rpc.GetAccounts(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetState() (handler http.Handler, ok bool) {
	values, err := url.ParseQuery(api.request.URL.RawQuery)
	if err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	// Params
	request := &types.SingleBytes{}
	account := values.Get("account")
	if account != "" {
		request.Value = []byte(account)
	}

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}

	return commonResponseHandler(api.rpc.GetState(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetStateAndProof() (handler http.Handler, ok bool) {
	request := &types.AccountAndRoot{}
	return commonResponseHandler(api.rpc.GetStateAndProof(api.request.Context(), request)), true
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

	return commonResponseHandler(api.rpc.GetNameInfo(api.request.Context(), request)), true
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
		hashBytes, err := base64.StdEncoding.DecodeString(hash)
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

	return commonResponseHandler(api.rpc.GetBlock(api.request.Context(), request)), true
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
	
	return commonResponseHandler(&types.BlockchainStatus{
		BestBlockHash:   last.BlockHash(),
		BestHeight:      last.GetHeader().GetBlockNo(),
		ConsensusInfo:   ca.GetConsensusInfo(),
		BestChainIdHash: bestChainIDHash,
		ChainInfo:       chainInfo,
	}, nil), true
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
		hashBytes, err := base64.StdEncoding.DecodeString(hash)
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

	return commonResponseHandler(api.rpc.GetBlockBody(api.request.Context(), request)), true
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

	return commonResponseHandler(api.rpc.ListBlockHeaders(api.request.Context(), request)), true
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
		hashBytes, err := base64.StdEncoding.DecodeString(hash)
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
	
	return commonResponseHandler(api.rpc.GetBlockMetadata(api.request.Context(), request)), true
}

func (api *Web3APIv1) ListBlockMetadata() (handler http.Handler, ok bool) {
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
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

	// Validate
	if _, err := govalidator.ValidateStruct(request); err != nil {
		return commonResponseHandler(&types.Empty{}, err), true
	}
	return commonResponseHandler(api.rpc.GetReceipt(api.request.Context(), request)), true
}

func (api *Web3APIv1) GetTX() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetBlockTX() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) VerifyTX() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) QueryContract() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) ListEvents() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetABI() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetAccountVotes() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) QueryContractState() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) NodeState() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetPeers() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetServerInfo() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetStaking() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetVotes() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetEnterpriseConfig() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}
func (api *Web3APIv1) GetConfChangeProgress() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}

func (api *Web3APIv1) CommitTX() (handler http.Handler, ok bool)	{
	request := &types.Empty{}
	return commonResponseHandler(request, status.Errorf(codes.Unknown, "Preparing")), true
}

