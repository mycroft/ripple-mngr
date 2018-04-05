package main

import (
	"crypto/rand"
	"math/big"

	"github.com/rubblelabs/ripple/crypto"
)

func newAccountId(key crypto.Key, sequence *uint32, useEd25519 bool) (crypto.Hash, error) {
	if useEd25519 {
		return crypto.AccountId(key, nil)
	}
	return crypto.AccountId(key, sequence)
}

func newKey(seed crypto.Hash, useEd25519 bool) (crypto.Key, error) {
	if useEd25519 {
		return crypto.NewEd25519Key(seed.Payload())
	}
	return crypto.NewECDSAKey(seed.Payload())
}

func GenerateKey() (crypto.Key, crypto.Hash, error) {
	useEd25519 := false
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return nil, nil, err
	}

	seed, err := crypto.NewFamilySeed(b)
	if err != nil {
		return nil, nil, err
	}

	key, err := newKey(seed, useEd25519)
	if err != nil {
		return nil, nil, err
	}

	return key, seed, err
}

func GetAddress(key crypto.Key) (string, error) {
	var sequenceZero uint32
	useEd25519 := false

	address, err := newAccountId(key, &sequenceZero, useEd25519)
	if err != nil {
		return "", err
	}

	return address.String(), nil
}

type AccountInfoParams struct {
	Account string `json:"account"`
}

type AccountData struct {
	Account string `json:"Account"`
	Balance string `json:"Balance"`
}

type InfoResult struct {
	AccountData AccountData `json:"account_data"`
}

func QueryBalance(account string) (big.Int, error) {
	var value big.Int
	var r Request
	r.Method = "account_info"
	r.Params = []AccountInfoParams{AccountInfoParams{account}}

	var infoResult InfoResult

	_, err := Query(r, &infoResult)
	if err != nil {
		return big.Int{}, err
	}

	value.SetString(infoResult.AccountData.Balance, 10)

	return value, nil
}

type TX struct {
	Account         string `json:"Account"`
	Amount          string `json:"Amount"`
	Destination     string `json:"Destination"`
	Fee             string `json:"Fee"`
	Flags           uint32 `json:"Flags`
	Sequence        uint32 `json:"Sequence"`
	SigningPubKey   string `json:"SigningPubKey"`
	TransactionType string `json:"TransactionType"`
	Hash            string `json:"hash"`
	Date            uint32 `json:"date"`
	InLedger        uint32 `json:"inLedger"`
	LedgerIndex     uint32 `json:"ledger_index"`
}

type Transactions struct {
	Tx        TX   `json:"tx"`
	Validated bool `json:"validated"`
}

type TxResult struct {
	Account      string         `json:"account"`
	Transactions []Transactions `json:"transactions"`
}

type Request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

func QueryTx(account string) ([]TX, error) {
	var tx []TX
	var r Request
	r.Method = "account_tx"
	r.Params = []AccountInfoParams{AccountInfoParams{account}}

	var txResult TxResult

	_, err := Query(r, &txResult)
	if err != nil {
		return []TX{}, err
	}

	for _, transaction := range txResult.Transactions {
		tx = append(tx, transaction.Tx)
	}

	return tx, err
}
