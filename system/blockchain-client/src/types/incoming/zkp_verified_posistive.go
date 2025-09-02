package incoming

type ZkpVerifiedPositiveDto struct {
	EventId        string        `json:"event_id"`
	ResolvedValues []ZkpFieldDto `json:"resolved_values"`
}
