package main

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		arg  *hardforkData
		want string
	}{
		{
			"dup version",
			&hardforkData{
				[]hardforkElem{
					{
						2, 100, 100,
					},
					{
						2, 200, 200,
					},
				},
				3,
				"test",
				"",
				"",
			},
			"version 3 expected, but got 2",
		},
		{
			"inverted mainnet block number",
			&hardforkData{
				[]hardforkElem{
					{
						2, 200, 100,
					},
					{
						3, 100, 200,
					},
				},
				3,
				"test",
				"",
				"",
			},
			"version 3, mainnet block number 100 is too low",
		},
		{
			"inverted testnet block number",
			&hardforkData{
				[]hardforkElem{
					{
						2, 200, 200,
					},
					{
						3, 200, 100,
					},
				},
				3,
				"test",
				"",
				"",
			},
			"version 3, testnet block number 100 is too low",
		},
		{
			"same block number",
			&hardforkData{
				[]hardforkElem{
					{
						2, 100, 100,
					},
					{
						3, 100, 100,
					},
				},
				3,
				"test",
				"",
				"",
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validate(tt.arg)
			if got != nil && got.Error() != tt.want {
				t.Errorf("validate() = %v, want: %v", got, tt.want)
			}
			if got == nil && len(tt.want) != 0 {
				t.Errorf("validate() has no error, want: %v", tt.want)
			}
		})
	}
}
