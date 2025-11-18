package zkprequest

// CircuitRegistry abstracts how you resolve schemas and keys for a circuit id.
type CircuitRegistry interface {
	// Return schema JSON for a given circuitID and whether it exists.
	SchemaJSON(circuitID string) (string, bool)

	// (Prod) Return verifier-held VK bytes for a circuitID (if you store VK server-side).
	VerifyingKey(circuitID string) ([]byte, bool)
}

// StaticRegistry is a simple in-memory registry (handy for dev/tests).
type StaticRegistry struct {
	Schemas map[string]string // circuitID -> schemaJSON
	VKs     map[string][]byte // circuitID -> verifying key bytes (optional)
}

func (r *StaticRegistry) SchemaJSON(circuitID string) (string, bool) {
	if r == nil || r.Schemas == nil {
		return "", false
	}
	v, ok := r.Schemas[circuitID]
	return v, ok
}

func (r *StaticRegistry) VerifyingKey(circuitID string) ([]byte, bool) {
	if r == nil || r.VKs == nil {
		return nil, false
	}
	v, ok := r.VKs[circuitID]
	return v, ok
}
