package types

type ZeroKnowledgeProofVerificationRequest struct {
	IdentityId string `json:"identity_id"`
	Schema     string `json:"schema"` // schema as JSON string
}
