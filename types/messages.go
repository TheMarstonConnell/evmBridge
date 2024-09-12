package types

type ExecuteMsg struct {
	PostKey  *ExecuteMsgPostKey  `json:"post_key,omitempty"`
	PostFile *ExecuteMsgPostFile `json:"post_file,omitempty"`
}

type ExecuteMsgPostKey struct {
	Key string `json:"key"`
}

type ExecuteMsgPostFile struct {
	Merkle        string `json:"merkle"`
	FileSize      int64  `json:"file_size"`
	ProofInterval int64  `json:"proof_interval"`
	ProofType     int64  `json:"proof_type"`
	MaxProofs     int64  `json:"max_proofs"`
	Expires       int64  `json:"expires"`
	Note          string `json:"note"`
}

// ToString returns a string representation of the message
func (m *ExecuteMsg) ToString() string {
	return toString(m)
}
