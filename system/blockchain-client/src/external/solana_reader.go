package external

import (
	"blockchain-client/src/zkp"
	"context"
	"net/http"
	"strconv"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-gonic/gin"
)

type SolanaReader struct {
	RpcClient *rpc.Client
}

type ZkpVerifier interface {
	Verify(c *gin.Context)
}

func NewSolanaReader() *SolanaReader {
	return &SolanaReader{
		RpcClient: rpc.New("http://host.docker.internal:8899"),
	}
}

// Verify godoc
// @Summary      Verify zkSNARK proof from Solana account
// @Description  Retrieves transaction and account data from Solana, extracts zkSNARK proof and verifies it
// @Tags         zkSNARK
// @Accept       json
// @Produce      json
// @Param        signature query string true "Transaction signature"
// @Param        account   query string true "Account public key"
// @Param        size      query int    true "Proof size in bytes"
// @Success      200 {object} map[string]interface{} "Successfully verified proof"
// @Failure      400 {object} map[string]string "Invalid request: missing parameters"
// @Failure      422 {object} map[string]string "Could not parse input"
// @Failure      404 {object} map[string]string "Transaction or account not found"
// @Failure      500 {object} map[string]string "Internal server error or verification failed"
// @Router       /verify [get]
func (sr *SolanaReader) Verify(c *gin.Context) {
	signature := c.Query("signature")
	accountStr := c.Query("account")
	proofSize := c.Query("size")

	if signature == "" || proofSize == "" || accountStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: signature, account, or size param is empty",
		})
		return
	}

	proofLen, err := strconv.Atoi(proofSize)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Could not parse proof size: " + err.Error(),
		})
		return
	}

	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Could not parse signature: " + err.Error(),
		})
		return
	}

	accountKey, err := solana.PublicKeyFromBase58(accountStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Could not parse account pubkey: " + err.Error(),
		})
		return
	}

	ctx := context.Background()

	_, err = sr.RpcClient.GetTransaction(
		ctx,
		sig,
		&rpc.GetTransactionOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Could not retrieve transaction from solana: " + err.Error(),
		})
		return
	}

	accInfo, err := sr.RpcClient.GetAccountInfo(ctx, accountKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Could not fetch account info: " + err.Error(),
		})
		return
	}
	if accInfo.Value == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	data := accInfo.Value.Data.GetBinary()
	if len(data) < proofLen {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Account data too short: expected at least " + strconv.Itoa(proofLen) +
				" bytes, got " + strconv.Itoa(len(data)),
		})
		return
	}

	zkpData := data[:proofLen]

	proof, err := zkp.ReconstructZkpResult(zkpData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to deserialize proof: " + err.Error(),
		})
		return
	}

	err = groth16.Verify(proof.Proof, proof.VerifyingKey, proof.PublicWitness)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify proof: " + err.Error()},
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"proof":   proof,
		"account": accountStr,
		"size":    proofLen,
	})
}
