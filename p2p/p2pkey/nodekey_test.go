/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pkey

import "testing"

func TestNodeVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeVersion(); got != tt.want {
				t.Errorf("NodeVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}