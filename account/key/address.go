/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"crypto/ecdsa"

	delegate "github.com/aergoio/aergo/account/key/crypto"
)

// GenerateAddress calculates the raw (not-encoded) address for a private key.
// Address is the raw, internally used identifier of an account
// In Aergo, the address is synonymous with the compressed public key (33 bytes).
func GenerateAddress(pubkey *ecdsa.PublicKey) []byte {
	return delegate.GenerateAddress(pubkey)
}
