package identity

import "api/src/model"

// mock validation
func validateSchema(schema model.Schema, request model.ZeroKnowledgeProofVerificationRequest) bool {
	return true
}
