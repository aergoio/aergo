/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package metric

// DataMetric hold data transfer of a peer. It shows eventually consistent aps and load score.
type DataMetric interface {
	// APS shows average bytes per second in a given period
	APS() int64
	// LoadScore is calulated value of network load.
	LoadScore() int64
	// add bytes transferred
	AddBytes(byteSize int)
	// Calculate should be called peridically with exact interval and a single thread
	Calculate()
}
