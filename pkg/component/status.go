/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

// StatusToString returns a string representation of a component's status
func StatusToString(status Status) string {
	switch status {
	case StartedStatus:
		return "started"
	case StoppingStatus:
		return "stopping"
	case StoppedStatus:
		return "stopped"
	case RestartingStatus:
		return "restarting"
	default:
		return "unknown"
	}
}
