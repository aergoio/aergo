/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package config

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	"github.com/spf13/viper"
)

func TestIsFork(t *testing.T) {
	type args struct {
		bno1 uint64
		bno2 uint64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"greater",
			args{
				10,
				14,
			},
			true,
		},
		{
			"equal",
			args{
				10,
				14,
			},
			true,
		},
		{
			"less",
			args{
				14,
				10,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFork(tt.args.bno1, tt.args.bno2); got != tt.want {
				t.Errorf("isFork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigFromToml(t *testing.T) {
	cfg := readConfig(`
[hardfork]
v2 = "9223"
v3 = "13000"
v4 = "14000"
v5 = "15000"
`,
	)
	if cfg.V2 != 9223 {
		t.Errorf("V2 = %d, want %d", cfg.V2, 9223)
	}
	if cfg.V3 != 13000 {
		t.Errorf("V3 = %d, want %d", cfg.V3, 13000)
	}
	if cfg.V4 != 14000 {
		t.Errorf("V4 = %d, want %d", cfg.V4, 14000)
	}
	if cfg.V5 != 15000 {
		t.Errorf("V5 = %d, want %d", cfg.V5, 15000)
	}
}

func TestCompatibility(t *testing.T) {
	cfg := readConfig(`
[hardfork]
v2 = "9223"
v3 = "11000"
v4 = "14000"
v5 = "15000"
`,
	)
	dbCfg, _ := readDbConfig(`
{
	"V2": 18446744073709551315,
	"V3": 18446744073709551415,
	"V4": 18446744073709551515,
	"V5": 18446744073709551615
}`,
	)
	err := cfg.CheckCompatibility(dbCfg, 10)
	if err != nil {
		t.Error(err)
	}

	dbCfg, _ = readDbConfig(`
{
	"V2": 9223,
	"V3": 10000,
	"V4": 14000,
	"V5": 15000
}`,
	)
	err = cfg.CheckCompatibility(dbCfg, 10)
	if err != nil {
		t.Error(err)
	}

	dbCfg, _ = readDbConfig(`
{
	"V2": 9223,
	"V3": 10000,
	"V4": 14000,
	"V5": 15000
}`,
	)
	err = cfg.CheckCompatibility(dbCfg, 9500)
	if err != nil {
		t.Error(err)
	}

	dbCfg, _ = readDbConfig(`
{
	"V2": 9221,
	"V3": 10000,
	"V4": 14000,
	"V5": 15000
}`,
	)
	err = cfg.CheckCompatibility(dbCfg, 9500)
	if err == nil {
		t.Error(`the expected error: the fork "V2" is incompatible: latest block(9500), node(9223), and chain(9221)`)
	}

	dbCfg, _ = readDbConfig(`
{
	"V2": 9223,
	"V3": 10000,
	"V4": 14000,
	"V5": 15000
}`,
	)
	err = cfg.CheckCompatibility(dbCfg, 10000)
	if err == nil {
		t.Error(`the expected error: the fork "V3" is incompatible: latest block(10000), node(0), and chain(10000)`)
	}

	dbCfg, _ = readDbConfig(`
{
	"V2": 9223,
	"V3": 10000
}`,
	)
	err = cfg.CheckCompatibility(dbCfg, 10001)
	if err == nil {
		t.Error(`the expected error: the fork "V3" is incompatible: latest block(10000), node(0), and chain(10000)`)
	}

	dbCfg, _ = readDbConfig(`
{
	"V2": 9223,
	"VV": 10000,
	"V3": 11000,
	"V4": 12000,
	"V5": 15000
}`,
	)
	err = cfg.CheckCompatibility(dbCfg, 9000)
	if err == nil {
		t.Error(`the expected error: strconv.ParseUint: parsing "V": invalid syntax`)
	}
	if _, ok := err.(*forkError); ok {
		t.Error(err)
	}
}

func TestVersion(t *testing.T) {
	cfg := readConfig(`
[hardfork]
v2 = "9223"
v3 = "10000"
v4 = "14000"
v5 = "20000"
`,
	)
	tests := []struct {
		name string
		h    uint64
		want int32
	}{
		{
			"zero",
			0,
			0,
		},
		{
			"less v2",
			10,
			0,
		},
		{
			"equal v2",
			9223,
			2,
		},
		{
			"greater v2",
			9322,
			2,
		},
		{
			"before v3",
			9999,
			2,
		},
		{
			"equal v3",
			10000,
			3,
		},
		{
			"greater v3",
			10001,
			3,
		},
		{
			"before v4",
			13999,
			3,
		},
		{
			"equal v4",
			14000,
			4,
		},
		{
			"greater v4",
			14001,
			4,
		},
		{
			"before v5",
			19999,
			4,
		},
		{
			"equal v5",
			20000,
			5,
		},
		{
			"greater v5",
			21001,
			5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.Version(tt.h); got != tt.want {
				t.Errorf("Version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFixDbConfig(t *testing.T) {
	cfg := readConfig(`
[hardfork]
v2 = "9223"
v3 = "10000"
v4 = "14000"
v5 = "20000"
`,
	)
	dbConfig, _ := readDbConfig(`
{
	"V2": 9223,
	"V3": 0
}`,
	)
	if err := cfg.CheckCompatibility(dbConfig, 10000); err == nil {
		t.Error("must be incompatible before fix")
	}
	dbConfig.FixDbConfig(*cfg)
	if err := cfg.CheckCompatibility(dbConfig, 10000); err == nil {
		t.Errorf("must be incompatible for existing height: %v", err)
	}

	dbConfig, _ = readDbConfig(`
{
	"V2": 9223
}`,
	)
	dbConfig.FixDbConfig(*cfg)
	if err := cfg.CheckCompatibility(dbConfig, 10000); err != nil {
		t.Errorf("must be compatible after fix for non-existing height: %v", err)
	}
}

func readConfig(c string) *HardforkConfig {
	v := viper.New()
	v.SetConfigType("toml")
	if err := v.ReadConfig(bytes.NewBuffer([]byte(c))); err != nil {
		log.Fatal(err)
	}
	cfg := new(Config)
	if err := v.Unmarshal(cfg); err != nil {
		log.Fatal(err)
	}
	return cfg.Hardfork
}

func readDbConfig(c string) (HardforkDbConfig, error) {
	var cfg HardforkDbConfig
	if err := json.Unmarshal([]byte(c), &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
