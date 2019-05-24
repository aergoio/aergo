/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2putil

import "io"

// ReadToLen read bytes from io.Reader until bf is filled.
func ReadToLen(rd io.Reader, bf []byte) (int, error) {
	remain := len(bf)
	offset := 0
	for remain > 0 {
		read, err := rd.Read(bf[offset:])
		if err != nil || read == 0 {
			return offset, err
		}
		remain -= read
		offset += read
	}
	return offset, nil
}
