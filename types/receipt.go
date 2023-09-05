package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/internal/merkle"
	"github.com/golang/protobuf/jsonpb"
	"github.com/minio/sha256-simd"
	"github.com/willf/bloom"
)

const (
	successStatus = iota
	createdStatus
	errorStatus
	recreatedStatus
)

func NewReceipt(contractAddress []byte, status string, jsonRet string) *Receipt {
	return &Receipt{
		ContractAddress: contractAddress[:33],
		Status:          status,
		Ret:             jsonRet,
	}
}

func (r *Receipt) marshalBody(b *bytes.Buffer, isMerkle bool) error {
	l := make([]byte, 8)
	b.Write(r.ContractAddress)
	var status byte
	switch r.Status {
	case "SUCCESS":
		status = successStatus
	case "CREATED":
		status = createdStatus
	case "ERROR":
		status = errorStatus
	case "RECREATED":
		status = recreatedStatus
	default:
		return errors.New("unsupported status in receipt")
	}
	b.WriteByte(status)
	if !isMerkle || status != errorStatus {
		binary.LittleEndian.PutUint32(l[:4], uint32(len(r.Ret)))
		b.Write(l[:4])
		b.WriteString(r.Ret)
	}
	b.Write(r.TxHash)

	binary.LittleEndian.PutUint32(l[:4], uint32(len(r.FeeUsed)))
	b.Write(l[:4])
	b.Write(r.FeeUsed)
	binary.LittleEndian.PutUint32(l[:4], uint32(len(r.CumulativeFeeUsed)))
	b.Write(l[:4])
	b.Write(r.CumulativeFeeUsed)

	if len(r.Bloom) == 0 {
		b.WriteByte(0)
	} else {
		b.WriteByte(1)
		b.Write(r.Bloom)
	}
	binary.LittleEndian.PutUint32(l[:4], uint32(len(r.Events)))
	b.Write(l[:4])

	return nil
}

func (r *Receipt) marshalBodyV2(b *bytes.Buffer, isMerkle bool) error {
	l := make([]byte, 8)
	b.Write(r.ContractAddress)
	var status byte
	switch r.Status {
	case "SUCCESS":
		status = successStatus
	case "CREATED":
		status = createdStatus
	case "ERROR":
		status = errorStatus
	case "RECREATED":
		status = recreatedStatus
	default:
		return errors.New("unsupported status in receipt")
	}
	b.WriteByte(status)
	if !isMerkle || status != errorStatus {
		binary.LittleEndian.PutUint32(l[:4], uint32(len(r.Ret)))
		b.Write(l[:4])
		b.WriteString(r.Ret)
	}
	b.Write(r.TxHash)

	binary.LittleEndian.PutUint32(l[:4], uint32(len(r.FeeUsed)))
	b.Write(l[:4])
	b.Write(r.FeeUsed)
	binary.LittleEndian.PutUint32(l[:4], uint32(len(r.CumulativeFeeUsed)))
	b.Write(l[:4])
	b.Write(r.CumulativeFeeUsed)
	binary.LittleEndian.PutUint64(l[:], r.GasUsed)
	b.Write(l)

	if r.FeeDelegation {
		b.WriteByte(1)
	} else {
		b.WriteByte(0)
	}
	if len(r.Bloom) == 0 {
		b.WriteByte(0)
	} else {
		b.WriteByte(1)
		b.Write(r.Bloom)
	}
	binary.LittleEndian.PutUint32(l[:4], uint32(len(r.Events)))
	b.Write(l[:4])

	return nil
}

func (r *Receipt) marshalStoreBinary() ([]byte, error) {
	var b bytes.Buffer

	err := r.marshalBody(&b, false)
	if err != nil {
		return nil, err
	}
	for _, ev := range r.Events {
		evB, err := ev.marshalStoreBinary(r)
		if err != nil {
			return nil, err
		}
		b.Write(evB)
	}

	return b.Bytes(), nil
}

func (r *Receipt) marshalStoreBinaryV2() ([]byte, error) {
	var b bytes.Buffer

	err := r.marshalBodyV2(&b, false)
	if err != nil {
		return nil, err
	}
	for _, ev := range r.Events {
		evB, err := ev.marshalStoreBinary(r)
		if err != nil {
			return nil, err
		}
		b.Write(evB)
	}

	return b.Bytes(), nil
}

func (r *Receipt) unmarshalBody(data []byte) ([]byte, uint32) {
	r.ContractAddress = data[:33]
	status := data[33]
	switch status {
	case successStatus:
		r.Status = "SUCCESS"
	case createdStatus:
		r.Status = "CREATED"
	case errorStatus:
		r.Status = "ERROR"
	case recreatedStatus:
		r.Status = "RECREATED"
	}
	pos := uint32(34)
	l := binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.Ret = string(data[pos : pos+l])
	pos += l
	r.TxHash = data[pos : pos+32]
	pos += 32
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.FeeUsed = data[pos : pos+l]
	pos += l
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.CumulativeFeeUsed = data[pos : pos+l]
	pos += l
	bloomCheck := data[pos]
	pos += 1
	if bloomCheck == 1 {
		r.Bloom = data[pos : pos+BloomBitByte]
		pos += BloomBitByte
	}
	pos += l
	evCount := binary.LittleEndian.Uint32(data[pos:])

	return data[pos+4:], evCount
}

func (r *Receipt) unmarshalBodyV2(data []byte) ([]byte, uint32) {
	r.ContractAddress = data[:33]
	status := data[33]
	switch status {
	case successStatus:
		r.Status = "SUCCESS"
	case createdStatus:
		r.Status = "CREATED"
	case errorStatus:
		r.Status = "ERROR"
	case recreatedStatus:
		r.Status = "RECREATED"
	}
	pos := uint32(34)
	l := binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.Ret = string(data[pos : pos+l])
	pos += l
	r.TxHash = data[pos : pos+32]
	pos += 32
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.FeeUsed = data[pos : pos+l]
	pos += l
	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	r.CumulativeFeeUsed = data[pos : pos+l]
	pos += l
	ll := binary.LittleEndian.Uint64(data[pos:])
	r.GasUsed = ll
	pos += 8
	if data[pos] == 1 {
		r.FeeDelegation = true
	}
	pos += 1
	bloomCheck := data[pos]
	pos += 1
	if bloomCheck == 1 {
		r.Bloom = data[pos : pos+BloomBitByte]
		pos += BloomBitByte
	}
	pos += l
	evCount := binary.LittleEndian.Uint32(data[pos:])

	return data[pos+4:], evCount
}

func (r *Receipt) unmarshalStoreBinary(data []byte) ([]byte, error) {
	evData, evCount := r.unmarshalBody(data)

	r.Events = make([]*Event, evCount)
	var err error
	for i := uint32(0); i < evCount; i++ {
		var ev Event
		evData, err = ev.unmarshalStoreBinary(evData, r)
		if err != nil {
			return nil, err
		}
		r.Events[i] = &ev
	}
	return evData, nil
}

func (r *Receipt) unmarshalStoreBinaryV2(data []byte) ([]byte, error) {
	evData, evCount := r.unmarshalBodyV2(data)

	r.Events = make([]*Event, evCount)
	var err error
	for i := uint32(0); i < evCount; i++ {
		var ev Event
		evData, err = ev.unmarshalStoreBinary(evData, r)
		if err != nil {
			return nil, err
		}
		r.Events[i] = &ev
	}
	return evData, nil
}

func (r *Receipt) MarshalBinaryTest() ([]byte, error) {
	return r.marshalStoreBinaryV2()
}

func (r *Receipt) MarshalMerkleBinary() ([]byte, error) {
	var b bytes.Buffer

	err := r.marshalBody(&b, true)
	if err != nil {
		return nil, err
	}

	for _, ev := range r.Events {
		evB, err := ev.MarshalMerkleBinary()
		if err != nil {
			return nil, err
		}
		b.Write(evB)
	}

	return b.Bytes(), nil
}

func (r *Receipt) MarshalMerkleBinaryV2() ([]byte, error) {
	var b bytes.Buffer

	err := r.marshalBodyV2(&b, true)
	if err != nil {
		return nil, err
	}

	for _, ev := range r.Events {
		evB, err := ev.MarshalMerkleBinary()
		if err != nil {
			return nil, err
		}
		b.Write(evB)
	}

	return b.Bytes(), nil
}

func (r *Receipt) UnmarshalBinaryTest(data []byte) error {
	_, err := r.unmarshalStoreBinaryV2(data)
	return err
}

func (r *Receipt) ReadFrom(data []byte) ([]byte, error) {
	evData, evCount := r.unmarshalBody(data)

	r.Events = make([]*Event, evCount)
	var err error
	for i := uint32(0); i < evCount; i++ {
		var ev Event
		evData, err = ev.UnmarshalBinary(evData)
		if err != nil {
			return nil, err
		}
		r.Events[i] = &ev
	}
	return evData, nil
}

func (r *Receipt) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	return json.Marshal(r)
}

func (r *Receipt) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"BlockNo":`)
	b.WriteString(fmt.Sprintf("%d", r.BlockNo))
	b.WriteString(`,"BlockHash":"`)
	b.WriteString(enc.ToString(r.BlockHash))
	b.WriteString(`","contractAddress":"`)
	b.WriteString(EncodeAddress(r.ContractAddress))
	b.WriteString(`","status":"`)
	b.WriteString(r.Status)
	if len(r.Ret) == 0 {
		b.WriteString(`","ret": {}`)
	} else if r.Status == "ERROR" {
		js, _ := json.Marshal(r.Ret)
		b.WriteString(`","ret": `)
		b.WriteString(string(js))
	} else {
		b.WriteString(`","ret": `)
		b.WriteString(r.Ret)
	}
	b.WriteString(`,"txHash":"`)
	b.WriteString(enc.ToString(r.TxHash))
	b.WriteString(`","txIndex":`)
	b.WriteString(fmt.Sprintf("%d", r.TxIndex))
	b.WriteString(`,"from":"`)
	b.WriteString(EncodeAddress(r.From))
	b.WriteString(`","to":"`)
	b.WriteString(EncodeAddress(r.To))
	b.WriteString(`","feeUsed":`)
	b.WriteString(new(big.Int).SetBytes(r.FeeUsed).String())
	b.WriteString(`,"gasUsed":`)
	b.WriteString(strconv.FormatUint(r.GasUsed, 10))
	b.WriteString(`,"feeDelegation":`)
	if r.FeeDelegation {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
	b.WriteString(`,"events":[`)
	for i, ev := range r.Events {
		if i != 0 {
			b.WriteString(`,`)
		}
		bEv, err := ev.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.Write(bEv)
	}
	b.WriteString(`]}`)
	return b.Bytes(), nil
}

func (r *Receipt) GetHash() []byte {
	h := sha256.New()
	b, _ := r.MarshalMerkleBinary()
	h.Write(b)
	return h.Sum(nil)
}

func (r *Receipt) BloomFilter(fi *FilterInfo) bool {
	if r.Bloom == nil {
		return false
	}

	var buffer bytes.Buffer
	l := make([]byte, 8)
	binary.BigEndian.PutUint64(l, BloomBitBits)
	buffer.Write(l)
	binary.BigEndian.PutUint64(l, BloomHashKNum)
	buffer.Write(l)
	binary.BigEndian.PutUint64(l, BloomBitBits)
	buffer.Write(l)
	buffer.Write(r.Bloom)

	var bf bloom.BloomFilter
	_, err := bf.ReadFrom(&buffer)
	if err != nil {
		return true
	}

	if bf.Test(fi.ContractAddress) || bf.Test([]byte(fi.EventName)) {
		return true
	}
	return false
}

func (r *Receipt) SetMemoryInfo(blkHash []byte, blkNo BlockNo, txIdx int32) {
	r.BlockNo = blkNo
	r.BlockHash = blkHash
	r.TxIndex = txIdx

	for _, e := range r.Events {
		e.SetMemoryInfo(r, blkHash, blkNo, txIdx)
	}
}

type bloomFilter bloom.BloomFilter

func (bf *bloomFilter) GetHash() []byte {
	h := sha256.New()
	b, _ := ((*bloom.BloomFilter)(bf)).GobEncode()
	h.Write(b)
	return h.Sum(nil)
}

type ReceiptMerkle struct {
	receipt        *Receipt
	blockNo        BlockNo
	hardForkConfig BlockVersionner
}

func (rm *ReceiptMerkle) GetHash() []byte {
	h := sha256.New()
	var b []byte
	if rm.hardForkConfig.IsV2Fork(rm.blockNo) {
		b, _ = rm.receipt.MarshalMerkleBinaryV2()
	} else {
		b, _ = rm.receipt.MarshalMerkleBinary()
	}
	h.Write(b)
	return h.Sum(nil)
}

type Receipts struct {
	bloom          *bloomFilter
	receipts       []*Receipt
	blockNo        BlockNo
	hardForkConfig BlockVersionner
}

func (rs *Receipts) Get() []*Receipt {
	if rs == nil {
		return nil
	}
	return rs.receipts
}

func (rs *Receipts) Set(receipts []*Receipt) {
	rs.receipts = receipts
}

func (rs *Receipts) SetHardFork(hardForkConfig BlockVersionner, blockNo BlockNo) {
	rs.hardForkConfig = hardForkConfig
	rs.blockNo = blockNo
}

const BloomBitByte = 256
const BloomBitBits = BloomBitByte * 8
const BloomHashKNum = 3

func (rs *Receipts) MergeBloom(bf *bloom.BloomFilter) error {
	if rs.bloom == nil {
		rs.bloom = (*bloomFilter)(bloom.New(BloomBitBits, BloomHashKNum))
	}

	return (*bloom.BloomFilter)(rs.bloom).Merge(bf)
}

func (rs *Receipts) BloomFilter(fi *FilterInfo) bool {
	if rs.bloom == nil {
		return false
	}
	bf := (*bloom.BloomFilter)(rs.bloom)
	if bf.Test(fi.ContractAddress) || bf.Test([]byte(fi.EventName)) {
		return true
	}
	return false
}

func (rs *Receipts) MerkleRoot() []byte {
	if rs == nil {
		return merkle.CalculateMerkleRoot(nil)
	}
	rsSize := len(rs.receipts)
	if rs.bloom != nil {
		rsSize++
	}
	mes := make([]merkle.MerkleEntry, rsSize)
	for i, r := range rs.receipts {
		mes[i] = &ReceiptMerkle{r, rs.blockNo, rs.hardForkConfig}
	}
	if rs.bloom != nil {
		mes[rsSize-1] = rs.bloom
	}
	return merkle.CalculateMerkleRoot(mes)
}

func (rs *Receipts) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	l := make([]byte, 4)

	if rs.bloom != nil {
		b.WriteByte(1)
		bloomB, err := (*bloom.BloomFilter)(rs.bloom).GobEncode()
		if err != nil {
			return nil, err
		}
		b.Write(bloomB[24:])
	} else {
		b.WriteByte(0)
	}
	binary.LittleEndian.PutUint32(l, uint32(len(rs.receipts)))
	b.Write(l)
	var rB []byte
	var err error
	for _, r := range rs.receipts {
		if rs.hardForkConfig.IsV2Fork(rs.blockNo) {
			rB, err = r.marshalStoreBinaryV2()
		} else {
			rB, err = r.marshalStoreBinary()
		}
		if err != nil {
			return nil, err
		}
		b.Write(rB)
	}

	return b.Bytes(), nil
}

func (rs *Receipts) UnmarshalBinary(data []byte) error {
	checkBloom := data[0]
	pos := 1
	if checkBloom == 1 {
		var buffer bytes.Buffer
		var bf bloom.BloomFilter
		l := make([]byte, 8)
		binary.BigEndian.PutUint64(l, BloomBitBits)
		buffer.Write(l)
		binary.BigEndian.PutUint64(l, BloomHashKNum)
		buffer.Write(l)
		binary.BigEndian.PutUint64(l, BloomBitBits)
		buffer.Write(l)
		buffer.Write(data[pos : pos+BloomBitByte])
		_, err := bf.ReadFrom(&buffer)
		if err != nil {
			return err
		}
		pos += BloomBitByte
		rs.bloom = (*bloomFilter)(&bf)
	}
	rCount := binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	rs.receipts = make([]*Receipt, rCount)
	unread := data[pos:]
	var err error
	for i := uint32(0); i < rCount; i++ {
		var r Receipt
		if rs.hardForkConfig.IsV2Fork(rs.blockNo) {
			unread, err = r.unmarshalStoreBinaryV2(unread)
		} else {
			unread, err = r.unmarshalStoreBinary(unread)
		}
		if err != nil {
			return err
		}
		rs.receipts[i] = &r
	}
	return nil
}

func (ev *Event) marshalCommonBinary(b *bytes.Buffer) {
	l := make([]byte, 4)
	b.Write(ev.ContractAddress)
	binary.LittleEndian.PutUint32(l, uint32(len(ev.EventName)))
	b.Write(l)
	b.WriteString(ev.EventName)

	binary.LittleEndian.PutUint32(l, uint32(len(ev.JsonArgs)))
	b.Write(l)
	b.WriteString(ev.JsonArgs)

	b.Write(ev.TxHash)

	binary.LittleEndian.PutUint32(l, uint32(ev.EventIdx))
	b.Write(l)
}

func (ev *Event) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	l := make([]byte, 8)
	ev.marshalCommonBinary(&b)

	b.Write(ev.BlockHash)
	binary.LittleEndian.PutUint64(l[:], ev.BlockNo)
	b.Write(l)
	binary.LittleEndian.PutUint32(l[:4], uint32(ev.TxIndex))
	b.Write(l[:4])
	return b.Bytes(), nil
}

func (ev *Event) marshalStoreBinary(r *Receipt) ([]byte, error) {
	var b bytes.Buffer
	l := make([]byte, 4)
	if bytes.Equal(r.ContractAddress, ev.ContractAddress) {
		b.WriteByte(0)
	} else {
		b.Write(ev.ContractAddress)
	}
	binary.LittleEndian.PutUint32(l, uint32(len(ev.EventName)))
	b.Write(l)
	b.WriteString(ev.EventName)

	binary.LittleEndian.PutUint32(l, uint32(len(ev.JsonArgs)))
	b.Write(l)
	b.WriteString(ev.JsonArgs)

	binary.LittleEndian.PutUint32(l, uint32(ev.EventIdx))
	b.Write(l)
	return b.Bytes(), nil
}

func (ev *Event) unmarshalStoreBinary(data []byte, r *Receipt) ([]byte, error) {
	var pos uint32
	if data[0] == 0 {
		ev.ContractAddress = r.ContractAddress
		pos += 1
	} else {
		ev.ContractAddress = data[:33]
		pos += 33
	}
	l := binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	ev.EventName = string(data[pos : pos+l])
	pos += l

	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	ev.JsonArgs = string(data[pos : pos+l])
	pos += l

	ev.EventIdx = int32(binary.LittleEndian.Uint32(data[pos:]))
	pos += 4

	return data[pos:], nil
}

func (ev *Event) MarshalMerkleBinary() ([]byte, error) {
	var b bytes.Buffer
	ev.marshalCommonBinary(&b)
	return b.Bytes(), nil
}

func (ev *Event) UnmarshalBinary(data []byte) ([]byte, error) {
	pos := uint32(33)
	ev.ContractAddress = data[:pos]

	l := binary.LittleEndian.Uint32(data[33:])
	pos += 4
	ev.EventName = string(data[pos : pos+l])
	pos += l

	l = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	ev.JsonArgs = string(data[pos : pos+l])
	pos += l

	ev.TxHash = data[pos : pos+32]
	pos += 32

	ev.EventIdx = int32(binary.LittleEndian.Uint32(data[pos:]))
	pos += 4

	ev.BlockHash = data[pos : pos+32]
	pos += 32

	ev.BlockNo = binary.LittleEndian.Uint64(data[pos:])
	pos += 8

	ev.TxIndex = int32(binary.LittleEndian.Uint32(data[pos:]))

	return data[pos+4:], nil
}

func (ev *Event) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"contractAddress":"`)
	b.WriteString(EncodeAddress(ev.ContractAddress))
	b.WriteString(`","eventName":"`)
	b.WriteString(ev.EventName)
	b.WriteString(`","Args":`)
	b.WriteString(ev.JsonArgs)
	b.WriteString(`,"txHash":"`)
	b.WriteString(enc.ToString(ev.TxHash))
	b.WriteString(`","EventIdx":`)
	b.WriteString(fmt.Sprintf("%d", ev.EventIdx))
	b.WriteString(`,"BlockHash":"`)
	b.WriteString(enc.ToString(ev.BlockHash))
	b.WriteString(`","BlockNo":`)
	b.WriteString(fmt.Sprintf("%d", ev.BlockNo))
	b.WriteString(`,"TxIndex":`)
	b.WriteString(fmt.Sprintf("%d", ev.TxIndex))
	b.WriteString(`}`)
	return b.Bytes(), nil
}

func (ev *Event) SetMemoryInfo(receipt *Receipt, blkHash []byte, blkNo BlockNo, txIdx int32) {
	ev.TxHash = receipt.TxHash
	ev.TxIndex = txIdx
	ev.BlockHash = blkHash
	ev.BlockNo = blkNo
}

func checkSameArray(value []interface{}, check []interface{}) bool {
	if len(value) != len(check) {
		return false
	}
	for i, v := range value {
		if checkValue(v, check[i]) == false {
			return false
		}
	}
	return true
}

func checkSameMap(value map[string]interface{}, check map[string]interface{}) bool {
	if len(value) != len(check) {
		return false
	}
	for k, v := range value {
		if checkValue(v, check[k]) == false {
			return false
		}
	}
	return true
}

func checkValue(value interface{}, check interface{}) bool {
	if reflect.TypeOf(value) != reflect.TypeOf(check) {
		return false
	}
	switch value.(type) {
	case string:
		if value.(string) != check.(string) {
			return false
		}
	case float64:
		if value.(float64) != check.(float64) {
			return false
		}
	case bool:
		if value.(bool) != check.(bool) {
			return false
		}
	case json.Number:
		if value.(json.Number) != check.(json.Number) {
			return false
		}
	case nil:
		return true
	case []interface{}:
		return checkSameArray(value.([]interface{}), check.([]interface{}))
	case map[string]interface{}:
		return checkSameMap(value.(map[string]interface{}), check.(map[string]interface{}))
	default:
		return false
	}

	return true
}

func (ev *Event) Filter(filter *FilterInfo, argFilter []ArgFilter) bool {
	if filter.ContractAddress != nil && !bytes.Equal(ev.ContractAddress, filter.ContractAddress) {
		return false
	}
	if len(filter.EventName) != 0 && ev.EventName != filter.EventName {
		return false
	}
	if argFilter != nil {
		var args []interface{}
		err := json.Unmarshal([]byte(ev.JsonArgs), &args)
		if err != nil {
			return false
		}
		argLen := len(args)
		for _, filter := range argFilter {
			if filter.argNo >= argLen {
				return false
			}
			value := args[filter.argNo]
			check := filter.value
			if checkValue(value, check) == false {
				return false
			}
		}
	}
	ev.ContractAddress = AddressOrigin(ev.ContractAddress)
	return true
}

type ArgFilter struct {
	argNo int
	value interface{}
}

const MAXBLOCKRANGE = 10000
const padprefix = 0x80

func AddressPadding(addr []byte) []byte {
	id := make([]byte, AddressLength)
	id[0] = padprefix
	copy(id[1:], addr)
	return id
}
func AddressOrigin(addr []byte) []byte {
	if addr[0] == padprefix {
		addr = addr[1:bytes.IndexByte(addr, 0)]
	}
	return addr
}

func (fi *FilterInfo) ValidateCheck(to uint64) error {
	if fi.ContractAddress == nil {
		return errors.New("invalid contractAddress:" + string(fi.ContractAddress))
	}
	if len(fi.ContractAddress) < AddressLength {
		fi.ContractAddress = AddressPadding(fi.ContractAddress)
	} else if len(fi.ContractAddress) != AddressLength {
		return errors.New("invalid contractAddress:" + string(fi.ContractAddress))
	}
	if fi.RecentBlockCnt > 0 {
		if fi.RecentBlockCnt > MAXBLOCKRANGE {
			return errors.New(fmt.Sprintf("too large value at recentBlockCnt %d (max %d)",
				fi.RecentBlockCnt, MAXBLOCKRANGE))
		}

	} else {
		if fi.Blockfrom+MAXBLOCKRANGE < to {
			return errors.New(fmt.Sprintf("too large block range(max %d) from %d to %d",
				MAXBLOCKRANGE, fi.Blockfrom, to))
		}
	}
	return nil
}

func (fi *FilterInfo) GetExArgFilter() ([]ArgFilter, error) {
	if len(fi.ArgFilter) == 0 {
		return nil, nil
	}

	var argMap map[string]interface{}
	err := json.Unmarshal(fi.ArgFilter, &argMap)
	if err != nil {
		return nil, errors.New("invalid json format:" + err.Error())
	}

	argFilter := make([]ArgFilter, len(argMap))
	i := 0
	for key, value := range argMap {
		idx, err := strconv.ParseInt(key, 10, 32)
		if err != nil || idx < 0 {
			return nil, errors.New("invalid argument number:" + key)
		}
		argFilter[i].argNo = int(idx)
		argFilter[i].value = value
		i++
	}
	if i > 0 {
		return argFilter[:i], nil
	}
	return nil, nil
}
