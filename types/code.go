package types

type Code interface {
	ByteCode() []byte
	ABI() []byte
	Len() int
	IsValidFormat() bool
	Bytes() []byte
}

type CodePayload interface {
	Code() Code
	HasArgs() bool
	Args() []byte
	Len() int
	IsValidFormat() (bool, error)
}
