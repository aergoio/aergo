/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	"errors"
	"strconv"
)

var RPCErrInvalidArgument = errors.New("invalid argument")
var RPCErrInternalError = errors.New("internal error")

func AddCategory(confs map[string]*ConfigItem, category string) *ConfigItem {
	cat := &ConfigItem{Props: make(map[string]string)}
	confs[category] = cat
	return cat
}

func (ci *ConfigItem) AddInt(key string, value int) *ConfigItem {
	ci.Add(key, strconv.Itoa(value))
	return ci
}

func (ci *ConfigItem) AddBool(key string, value bool) *ConfigItem {
	ci.Add(key, strconv.FormatBool(value))
	return ci
}

func (ci *ConfigItem) AddFloat(key string, value float64) *ConfigItem {
	ci.Add(key, strconv.FormatFloat(value, 'g', -1, 64))
	return ci
}

func (ci *ConfigItem) Add(key, value string) *ConfigItem {
	ci.Props[key] = value
	return ci
}
