package incoming

import "blockchain-client/src/types/domain"

type ZkpVerifiedPositiveDto struct {
	// TODO: mock structure replace with actual implmentation
	IdentityId string `json:"identity_id"`
	Day        int    `json:"day"`
	Month      int    `json:"month"`
	Year       int    `json:"year"`
	Schema     string `json:"schema"`
}

func (zvp ZkpVerifiedPositiveDto) MapToCircuitBase() domain.ZkpCircuitBase {
	// TODO: this is simplified date mapping need to map from actaul date type
	return domain.ZkpCircuitBase{
		Day:   zvp.Day,
		Month: zvp.Month,
		Year:  zvp.Year,
	}
}
