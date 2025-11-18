// zkp_storage_data.go
package zkp

import "github.com/gagliardetto/solana-go"

type ZkpStorageData struct {
	Account   solana.PublicKey
	Signature solana.Signature
}
