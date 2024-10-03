package utils

import (
	"encoding/hex"
	"strconv"
	"strings"
)

func StartsWith0x(input string) bool {
	return strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "0X")
}

func Remove0x(input string) string {
	return strings.Replace(input, "0x", "", -1)
}

func DecodeUint64(input string) (uint64, error) {
	input = Remove0x(input)
	return strconv.ParseUint(input, 16, 64)
}

func EncodeHex(v uint64) string {
	return "0x" + strconv.FormatUint(v, 16)
}

func NormalizeHex(s string) string {
	lower := strings.ToLower(s)
	return Remove0x(lower)
}

func DenormalizeHex(s string) string {
	return "0x" + s
}

func NormalizeBytesToHex(b []byte) string {
	return NormalizeHex(hex.EncodeToString(b))
}
