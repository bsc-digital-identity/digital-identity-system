package incoming

type ZkpVerifiedNegativeDto struct {
	// TODO: mock structure replace with actual implmentation
	IdentityId string `json:"identity_id"`
	Schema     string `json:"schema"`
	Reason     string `json:"reason"`
}
