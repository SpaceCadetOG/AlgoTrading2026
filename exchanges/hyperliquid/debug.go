package hyperliquid

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	hl "github.com/sonirico/go-hyperliquid"
)

type debugL1Signer struct {
	privateKey *ecdsa.PrivateKey
	client     *Client
}

func (s *debugL1Signer) SignL1Action(
	ctx context.Context,
	action any,
	vaultAddress string,
	timestamp int64,
	expiresAfter *int64,
	isMainnet bool,
) (hl.SignatureResult, error) {
	s.client.debugHLAction(actionType(action), &timestamp)
	return hl.SignL1Action(s.privateKey, action, vaultAddress, timestamp, expiresAfter, isMainnet)
}

func (c *Client) debugHLAction(actionType string, nonce *int64) {
	if os.Getenv("HL_DEBUG_SIGNING") != "true" {
		return
	}

	nonceText := "<library-managed>"
	if nonce != nil {
		nonceText = formatInt64(*nonce)
	}

	log.Printf("HL signing endpoint=%s/exchange actionType=%s nonce=%s recoveredSigner=%s", getBaseURL(), actionType, nonceText, c.signerAddress())
}

func (c *Client) debugHLResponse(actionType string, response any) {
	if os.Getenv("HL_DEBUG_SIGNING") != "true" {
		return
	}

	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("HL response actionType=%s marshalError=%v", actionType, err)
		return
	}

	log.Printf("HL response actionType=%s body=%s", actionType, string(body))
}

func debugHLREST(endpoint string, request string, response []byte) {
	if os.Getenv("HL_DEBUG_SIGNING") != "true" {
		return
	}

	log.Printf("HL REST endpoint=%s request=%s responseBody=%s", endpoint, request, string(response))
}

func (c *Client) signerAddress() string {
	privateKey, err := c.privateKey()
	if err != nil {
		return "<unavailable>"
	}

	return crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
}

func actionType(action any) string {
	if action == nil {
		return "<unknown>"
	}

	data, err := json.Marshal(action)
	if err == nil {
		var obj map[string]any
		if json.Unmarshal(data, &obj) == nil {
			if typ, ok := obj["type"].(string); ok && typ != "" {
				return typ
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(action))
	if value.IsValid() && value.Kind() == reflect.Struct {
		field := value.FieldByName("Type")
		if field.IsValid() && field.Kind() == reflect.String && field.String() != "" {
			return field.String()
		}
	}

	return reflect.TypeOf(action).String()
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}
