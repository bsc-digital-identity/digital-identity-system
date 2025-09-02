package reasoncodes

type ReasonCode string

const (
	ErrUnmarshal          ReasonCode = "UnmarshalError"
	ErrVerifierResolution ReasonCode = "VerifierResolutionError"
	ErrProofGeneration    ReasonCode = "ProofGenerationError"
	ErrSolana             ReasonCode = "SolanaBlockchainError"
)
