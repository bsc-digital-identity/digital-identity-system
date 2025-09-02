package incoming

type ZkpVerifiedNegativeDto struct {
	EventId        string        `json:"event_id"`
	ResolvedValues []ZkpFieldDto `json:"resolved_values"`
}
