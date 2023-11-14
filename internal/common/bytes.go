package common

// IsZero returns true if argument is empty or zero
func IsZero(argv []byte) bool {
	if len(argv) == 0 {
		return true
	}
	for i := range argv {
		if argv[i] != 0x00 {
			return false
		}
	}
	return true
}

// Compactz returns nil if argument is empty or zero
func Compactz(argv []byte) []byte {
	if IsZero(argv) {
		return nil
	}
	return argv
}
