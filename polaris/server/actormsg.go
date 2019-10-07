/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

const PolarisSvc = "polarisSvc"

type PaginationMsg struct {
	ReferenceHash []byte
	Size          uint32
}

type CurrentListMsg PaginationMsg
type WhiteListMsg PaginationMsg
type BlackListMsg PaginationMsg

type ListEntriesMsg struct {
}