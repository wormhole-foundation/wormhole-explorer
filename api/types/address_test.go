package types

import "testing"

// Test_Address_ShortString runs several test cases on the method `Address.ShortString()`.
func Test_Address_ShortString(t *testing.T) {

	testCases := []struct {
		Input    string
		Hex      string
		ShortHex string
	}{
		{
			Input:    "0x000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7",
			Hex:      "000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7",
			ShortHex: "f890982f9310df57d00f659cf4fd87e65aded8d7",
		},
		{
			Input:    "000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7",
			Hex:      "000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7",
			ShortHex: "f890982f9310df57d00f659cf4fd87e65aded8d7",
		},
		{
			Input:    "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
			Hex:      "000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7",
			ShortHex: "f890982f9310df57d00f659cf4fd87e65aded8d7",
		},
		{
			Input:    "f890982f9310df57d00f659cf4fd87e65aded8d7",
			Hex:      "000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7",
			ShortHex: "f890982f9310df57d00f659cf4fd87e65aded8d7",
		},
		{
			Input:    "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
			Hex:      "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
			ShortHex: "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
		},
		{
			Input:    "0xec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
			Hex:      "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
			ShortHex: "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
		},
	}

	for i := range testCases {
		tc := &testCases[i]

		addr, err := StringToAddress(tc.Input)
		if err != nil {
			t.Fatalf("failed to parse address %s: %v", tc.Input, err)
		}

		if addr.Hex() != tc.Hex {
			t.Fatalf("expected Address.Hex()=%s, but got %s", tc.Hex, addr.Hex())
		}

		if addr.ShortHex() != tc.ShortHex {
			t.Fatalf("expected Address.ShortHex()=%s, but got %s", tc.ShortHex, addr.ShortHex())
		}
	}

}
