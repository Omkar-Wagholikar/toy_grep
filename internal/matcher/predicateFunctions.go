package matcher

func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
