package vaa

// ProcessVaaRequest request a vaa to process.
type ProcessVaaRequest struct {
	ID string `json:"id"`
}

// TxHashRequest request a tx hash.
type TxHashRequest struct {
	VaaID  string `json:"id"`
	TxHash string `json:"txHash"`
}

// ProcessVaaResponse response from processing a vaa.
type TxHashResponse struct {
	NativeTxHash string `json:"nativeTxHash"`
}
