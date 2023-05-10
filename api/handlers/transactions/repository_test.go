package transactions

import "testing"

func Test_convertToDecimal(t *testing.T) {

	tcs := []struct {
		input  int64
		output string
	}{
		{
			input:  1,
			output: "0.00000001",
		},
		{
			input:  1000_0000,
			output: "0.10000000",
		},
		{
			input:  1_0000_0000,
			output: "1.00000000",
		},
		{
			input:  1234_5678_1234,
			output: "1234.56781234",
		},
	}

	for i := range tcs {
		tc := tcs[i]

		result := convertToDecimal(tc.input)
		if result != tc.output {
			t.Errorf("expected %s, got %s", tc.output, result)
		}
	}

}
