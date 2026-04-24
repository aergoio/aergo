# 

This version have a feature for NHIS only. 

It listen other two ports to control internal DB and cannot turn off by default.

You need to set environment variable to turn off this feature.
	
- `TURN_OFF_CONTROL_COMPACTION` : Turn off control compaction feature. Default is `false`
- `COMPACTION_CHAIN_DB_PORT` : Port for compaction chain DB. Default is `17092`
- `COMPACTION_STATE_DB_PORT` : Port for compaction state DB. Default is `17091`
