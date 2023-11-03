package contract

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc"
)

func TestVerify(t *testing.T) {
	tests := []struct {
		key    []byte
		value  []byte
		proof  [][]byte
		verify bool
	}{
		{
			toBytes("0xa6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49"),
			toBytes("0x2710"),
			proofToBytes([]string{
				"0xf871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080",
				"0xe5a03eb5be412f275a18f6e4d622aee4ff40b21467c926224771b782d4c095d1444b83822710",
			}),
			true,
		},
		{
			toBytes("0x0000000000000000000000000000000000000000000000000000000000000000"),
			toBytes("0x6b746c656500000000000000000000000000000000000000000000000000000a"),
			proofToBytes([]string{
				"0xf844a120290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563a1a06b746c656500000000000000000000000000000000000000000000000000000a",
			}),
			true,
		},
		{
			toBytes("0x0000000000000000000000000000000000000000000000000000000000000001"),
			toBytes("0x6b736c656500000000000000000000000000000000000000000000000000000a"),
			proofToBytes([]string{
				"0xf8518080a05d2423f2a53dd794285a39f2c7ba5c0c24cba719d05e3bdd1f5eefae445b34358080808080808080a01ec473dfa012cb440907fa4f2c34be3e733c92430d98a48831700bc8ab159f5d8080808080",
				"0xf843a0310e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6a1a06b736c656500000000000000000000000000000000000000000000000000000a",
			}),
			true,
		},
		{
			toBytes("0x0000000000000000000000000000000000000000000000000000000000000000"),
			toBytes("0x9"),
			proofToBytes([]string{
				"0xf8518080a05d2423f2a53dd794285a39f2c7ba5c0c24cba719d05e3bdd1f5eefae445b34358080808080808080a01ec473dfa012cb440907fa4f2c34be3e733c92430d98a48831700bc8ab159f5d8080808080",

				"0xe2a0390decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e56309",
			}),
			true,
		},
		{
			toBytes("0x0000000000000000000000000000000000000000000000000000000000000000"),
			toBytes("0x6b736c656500000000000000000000000000000000000000000000000000000a"),
			proofToBytes([]string{
				"0xf871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080",
				"0xf843a0390decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563a1a06b736c656500000000000000000000000000000000000000000000000000000a",
			}),
			true,
		},
		{
			toBytes("0xac33ff75c19e70fe83507db0d683fd3465c996598dc972688b7ace676c89077b"),
			toBytes("0x6b746c656500000000000000000000000000000000000000000000000000000a"),
			proofToBytes([]string{
				"0xf871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080",
				"0xf843a03d2944a272ac5bae96b5bd2f67b6c13276d541dc09eb1cf414d96b19a09e1c2fa1a06b746c656500000000000000000000000000000000000000000000000000000a",
			}),
			true,
		},
		{
			toBytes("ac33ff75c19e70fe83507db0d683fd3465c996598dc972688b7ace676c89077b"),
			toBytes("6b746c656500000000000000000000000000000000000000000000000000000a"),
			proofToBytes([]string{
				"0xf871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080",
				"0xf843a03d2944a272ac5bae96b5bd2f67b6c13276d541dc09eb1cf414d96b19a09e1c2fa1a06b746c656500000000000000000000000000000000000000000000000000000a",
			}),
			true,
		},
		{
			toBytes("0xac33ff75c19e70fe83507db0d683fd3465c996598dc972688b7ace676c89077b"),
			toBytes(""),
			proofToBytes([]string{
				"0xf871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080",
				"0xf843a03d2944a272ac5bae96b5bd2f67b6c13276d541dc09eb1cf414d96b19a09e1c2fa1a06b746c656500000000000000000000000000000000000000000000000000000a",
			}),
			false,
		},
	}

	for i, tt := range tests {
		if verifyEthStorageProof(tt.key, rlpString(tt.value), keccak256(tt.proof[0]), tt.proof) != tt.verify {
			t.Errorf("testcase %d: want %v, got %v\n", i, tt.verify, !tt.verify)
		}
	}
}

func TestDecodeRlpTrieNode(t *testing.T) {
	tests := []struct {
		data   []byte
		length int
		index  int
		expect []byte
	}{
		{
			[]byte{0xc5, 0x83, 'c', 'a', 't', 0x80},
			2,
			-1,
			[]byte{},
		},
		{
			[]byte{0xc8, 0x83, 'c', 'a', 't', 0x83, 'd', 'o', 'g'},
			2,
			-1,
			[]byte{},
		},
		{
			[]byte{0xC0},
			0,
			-1,
			[]byte{},
		},
		{
			[]byte{0xd1, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q'},
			17,
			-1,
			[]byte{},
		},
		{
			[]byte{0xd1, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r'},
			0,
			-1,
			[]byte{},
		},
		{
			[]byte{0xd1, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 0x80},
			17,
			-1,
			[]byte{},
		},
		{
			[]byte{0xd1, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 0x80},
			17,
			0,
			[]byte{'a'},
		},
		{
			[]byte{0xd1, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 0x80},
			17,
			16,
			[]byte{},
		},
		{
			[]byte{0xd3, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 0x82, 'm', 'n', 'o', 'p', 0x80, 'x'},
			17,
			12,
			[]byte{'m', 'n'},
		},
	}
	for i, tt := range tests {
		n := decodeRlpTrieNode(tt.data)
		if len(n) != tt.length {
			t.Errorf("testcase %d: want %d, got %d\n", i, tt.length, len(n))
			continue
		}
		if tt.index != -1 {
			if !bytes.Equal(tt.expect, n[tt.index]) {
				t.Errorf("testcase %d: want %v, got %v\n", i, tt.expect, n[tt.index])
			}
		}
	}
}

func TestKeccak256(t *testing.T) {
	tests := []struct {
		msg string
		h   string
	}{
		{
			"abc",
			"4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45",
		},
		{
			"aergo",
			"e98bb03ab37161f8bbfe131f711dcccf3002a9cd9ec31bbd52edf181f7ab09a0",
		},
	}
	for _, tt := range tests {
		h := keccak256Hex([]byte(tt.msg))
		if tt.h != h {
			t.Errorf("want %s, got %s\n", tt.h, h)
		}
	}
}

func removeHexPrefix(s string) string {
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}
	if len(s)&1 == 1 {
		s = "0" + s
	}
	return s
}

func toBytes(s string) []byte {
	n, _ := enc.HexDecode(removeHexPrefix(s))
	return n
}

func proofToBytes(proof []string) [][]byte {
	var r [][]byte
	for _, n := range proof {
		d, err := enc.HexDecode(removeHexPrefix(n))
		if err != nil {
			return [][]byte{}
		}
		r = append(r, d)
	}
	return r
}

func Test_rlpEncode(t *testing.T) {
	l := "Lorem ipsum dolor sit amet, consectetur adipisicing elit"
	type args struct {
		o rlpObject
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			`"dog"`,
			args{rlpString("dog")},
			[]byte{0x83, 'd', 'o', 'g'},
		},
		{
			`["cat","dog"]`,
			args{rlpList{rlpString("cat"), rlpString("dog")}},
			[]byte{0xc8, 0x83, 'c', 'a', 't', 0x83, 'd', 'o', 'g'},
		},
		{
			`""`,
			args{rlpString("")},
			[]byte{0x80},
		},
		{
			`[]`,
			args{rlpList{}},
			[]byte{0xc0},
		},
		{
			`\x00`,
			args{rlpString{0x00}},
			[]byte{0x00},
		},
		{
			`\x0f`,
			args{rlpString{0x0f}},
			[]byte{0x0f},
		},
		{
			`\x0f\x00`,
			args{rlpString{0x04, 0x00}},
			[]byte{0x82, 0x04, 0x00},
		},
		{
			`[[],[[]],[[],[[]]]]`,
			args{rlpList{rlpList{}, rlpList{rlpList{}}, rlpList{rlpList{}, rlpList{rlpList{}}}}},
			[]byte{0xc7, 0xc0, 0xc1, 0xc0, 0xc3, 0xc0, 0xc1, 0xc0},
		},
		{
			fmt.Sprintf(`"%s"`, l),
			args{rlpString(l)},
			append([]byte{0xb8, 0x38}, []byte(l)...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rlpEncode(tt.args.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rlpEncode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rlpLength(t *testing.T) {
	type args struct {
		dataLen int
		offset  byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"string-1024",
			args{1024, 0x80},
			[]byte{0xb9, 0x04, 0x00},
		},
		{
			"list-1024",
			args{1024, 0xc0},
			[]byte{0xf9, 0x04, 0x00},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rlpLength(tt.args.dataLen, tt.args.offset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rlpLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
