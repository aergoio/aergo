package enterprise

import (
	"os"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

var cdb *state.ChainStateDB
var sdb *state.StateDB

func initTest(t *testing.T) (*state.ContractState, *state.V, *state.V) {
	cdb = state.NewChainStateDB()
	cdb.Init(string(db.BadgerImpl), "test", nil, false)
	genesis := types.GetTestGenesis()
	sdb = cdb.OpenNewStateDB(cdb.GetRoot())
	err := cdb.SetGenesis(genesis, nil)
	if err != nil {
		t.Fatalf("failed init : %s", err.Error())
	}
	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"

	scs, err := cdb.GetStateDB().OpenContractStateAccount(types.ToAccountID([]byte("aergo.system")))
	assert.NoError(t, err, "could not open contract state")

	account, err := types.DecodeAddress(testSender)
	assert.NoError(t, err, "could not decode test address")
	sender, err := sdb.GetAccountStateV(account)
	assert.NoError(t, err, "could not get test address state")
	receiver, err := sdb.GetAccountStateV([]byte(types.AergoEnterprise))
	assert.NoError(t, err, "could not get test address state")
	return scs, sender, receiver
}

func deinitTest() {
	cdb.Close()
	os.RemoveAll("test")
}

func TestSetGetConf(t *testing.T) {
	scs, _, _ := initTest(t)
	defer deinitTest()
	testConf := &Conf{On: true, Values: []string{"abc:w", "def:r", "ghi:s"}}
	retConf, err := getConf(scs, []byte(RPCPermissions))
	assert.NoError(t, err, "could not get test conf")
	assert.Nil(t, retConf, "not set yet")
	err = setConf(scs, []byte(RPCPermissions), testConf)
	assert.NoError(t, err, "could not set test conf")
	retConf, err = getConf(scs, []byte(RPCPermissions))
	assert.NoError(t, err, "could not get test conf")
	assert.Equal(t, testConf.Values, retConf.Values, "check values")
	assert.Equal(t, testConf.On, retConf.On, "check on")

	testConf2 := &Conf{On: false, Values: []string{"1", "22", "333"}}
	err = setConf(scs, []byte("p2pwhite"), testConf2)
	assert.NoError(t, err, "could not set test conf")
	retConf2, err := getConf(scs, []byte("p2pwhite"))
	assert.NoError(t, err, "could not get test conf")
	assert.Equal(t, testConf2.Values, retConf2.Values, "check values")
	assert.Equal(t, testConf2.On, retConf2.On, "check on")
}

const testCert = `-----BEGIN CERTIFICATE-----
MIIE4DCCAsigAwIBAgIBATANBgkqhkiG9w0BAQsFADAQMQ4wDAYDVQQDEwVhZXJn
bzAeFw0xOTA1MjIwMTU3MzJaFw0yMDExMjIwMTU3MzJaMBAxDjAMBgNVBAMTBWFl
cmdvMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA6y/dza6BPg6l//Hv
+NR5bgtEVNG68W38vfn/M2WDA9PfIaf2nlH3RiawMpw4G6cT/+qGoESCypDLEASS
NUo1Xkl5Ms78rus3PM6wVLuA6qgtyBEFKOu6ChMcmNh2RtybA10hgyPxakDilSZs
ZnONPBMGEnBFYGPNZm/VZgsavuIcJNnTDbPFjXyiet3+xEflNDA8PPdEHZXkYRCS
Z1bDH+iMm37KbFyJ2i7UTVmP21OFYnoNHcZTuYmbgCm/w9Fvf2vkY5fl3OidFKse
MqCB4VwWbEXlCGarrFrhjZ2Vd6M6YCwcq3lX+wC3f7mFwRa39BWZs+VFW2KPhuvr
jKR1lWjtM6PIl0mhqhJjShzrb2wgX6RaSVrOs1E+dQL9vNB9hQieShUVyd4m433Y
+gpJHeNhVZ+yBuNzm8wB3BB9e4mf4pKB9WFiLO9Gdi35hnldbUne9F78IJXSeIsv
MmZIu3hHkHeHcsTXrK9AQuP2XTAxgZ6OjdlmvZ9kGNHxikAKH0reMSOU5IBn6FwX
+7uKoU/Gyj5jTruMbwfbkr5SsEg0cG+xIT8l+ml3V7kOm3xYFfPRBhFTJIMcqEf5
QD8OlLmOMvDQ/3+w+G+ZngvS/bos8vzMxkpDM/Dr4UY1OFYS0LLXppKUJEC3XJlj
ReGCs4H07gtZiVEW8hc1xJ9VuS0CAwEAAaNFMEMwDgYDVR0PAQH/BAQDAgEGMBIG
A1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0OBBYEFPbpyqXvBRGAXvDBgDNp/vyB++kH
MA0GCSqGSIb3DQEBCwUAA4ICAQDkqhA2LQC9PCbPdyi6gMpQG0ed8/RpTdInIbad
HaSdmh6p1AIUkVaN84SzL1Il+LxWV/pTsTKDjEGI9ii0+zfoMFqA1wjF8LerjVSz
qCXS++fZFIKU12wSR0wWwlcsnuck7xgqmayG8K1jQuoIe3E/TTpjyMy9mHABDH50
EgV5O1WpBLghO+vCZVDNuSkwrCyg/jdYC1x0uLDb5OXQdQ3NO+eha5bxHgHhJXVB
Ag4uv/uQlR00wnWC62DCp2Be3G4JA5y6kNA9aOoEya8Q1UEVaUBRmGczUOjT9mjE
jVfkcCbeKtOtDN9h6TIqQCTL7Bct/lBe3cK9zBVyLJH0SlEgDfJd4AU01O330X6E
wEOB9MCCq3skeYbfxHO5515S28946KEplzZ9uJQvq0euM34CIas2Y0o7YhbqqWpi
0114GKXCpAue4tY4hkTzEGjlIL6MnKeHaccyKZijtWyz0yuqUoxIvimTdJ9AlD++
+SCxaPCAnN2mRhZvvVFJNTnTJ67M+Ne9uaGxrXd3R0sip95rOBZPs/Snh63zurhi
ENJBQIZFGoQuNyxewNgeZ52pLHhFXgQXR14EyR0hvmya9PqIK7lcHtxl/9oe4mk5
JkNFaNRPJQ8r0mw4Xc0wwXc1rlwToVD/EY+VWSblHuk9S7wYX1wAHKGWvW6AaxFZ
08S1rA==
-----END CERTIFICATE-----`
