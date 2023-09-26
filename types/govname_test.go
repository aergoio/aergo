package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerializeNameMap(t *testing.T) {
	for _, test := range []*NameMap{
		{
			Version:     1,
			Destination: []byte(""),
			Owner:       []byte(""),
		},
		{
			Version:     1,
			Destination: []byte("dest"),
			Owner:       []byte("owner"),
		},
	} {
		data := SerializeNameMap(test)
		test2 := DeserializeNameMap(data)
		require.EqualValuesf(t, test, test2, "name map serialization failed %s %s", test, test2)
	}
}
