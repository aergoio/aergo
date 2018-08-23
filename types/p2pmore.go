package types

// GetMessageData is delegation method for backward compatability
func (m *P2PMessage) GetMessageData() *MessageData {
	return m.Header
}
