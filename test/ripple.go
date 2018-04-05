package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rubblelabs/ripple/crypto"
)

func printKeys(seed crypto.Hash, key crypto.Key, useEd25519 bool) error {
	fmt.Println("Seed (secret key)", seed)

	var sequenceZero uint32

	accountId, err := newAccountId(key, &sequenceZero, useEd25519)
	if err != nil {
		return err
	}

	fmt.Println("AccountID", accountId)

	return nil
}

func newKey(seed crypto.Hash, useEd25519 bool) (crypto.Key, error) {
	if useEd25519 {
		return crypto.NewEd25519Key(seed.Payload())
	}
	return crypto.NewECDSAKey(seed.Payload())
}

func newAccountId(key crypto.Key, sequence *uint32, useEd25519 bool) (crypto.Hash, error) {
	if useEd25519 {
		return crypto.AccountId(key, nil)
	}
	return crypto.AccountId(key, sequence)
}

func generateKeyPairRandom() error {
	useEd25519 := true
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return err
	}

	seed, err := crypto.NewFamilySeed(b)
	if err != nil {
		return err
	}

	key, err := newKey(seed, useEd25519)
	if err != nil {
		return err
	}

	err = printKeys(seed, key, useEd25519)
	if err != nil {
		return err
	}

	return nil
}

func generateKeyPairSeed(s string) error {
	seed, err := crypto.NewRippleHash(s)
	if err != nil {
		return err
	}

	key, err := newKey(seed, false)
	if err != nil {
		return err
	}

	err = printKeys(seed, key, false)
	if err != nil {
		return err
	}

	return nil
}

type AccountInfoParam struct {
	Account string `json:"account"`
}

type AccountInfo struct {
	Method string             `json:"method"`
	Params []AccountInfoParam `json:"params"`
}

type AccountTxParams struct {
	Account string `json:"account"`
}

type AccountTx struct {
	Method string            `json:"method"`
	Params []AccountTxParams `json:"params"`
}

type RequestResult struct {
	Result interface{} `json:"result"`
}

func Query(r interface{}, out interface{}) ([]byte, error) {
	jsonStr, err := json.Marshal(r)
	if err != nil {
		return []byte{}, err
	}

	fmt.Println(string(jsonStr))

	req, err := http.NewRequest("POST", "https://s.altnet.rippletest.net:51234", bytes.NewBuffer(jsonStr))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if out != nil {
		var rr RequestResult
		rr.Result = out

		err := json.Unmarshal(body, &rr)
		if err != nil {
			return body, err
		}
	}

	return body, err
}

func QueryBalance(account string) {
	var account_tx AccountTx
	account_tx.Method = "account_info"
	account_tx.Params = []AccountTxParams{AccountTxParams{account}}

	body, err := Query(account_tx, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
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

func QueryTx(account string) (TxResult, error) {
	var r Request
	r.Method = "account_tx"
	r.Params = []AccountTxParams{AccountTxParams{account}}

	var txResult TxResult

	_, err := Query(r, &txResult)
	if err != nil {
		return TxResult{}, err
	}

	return txResult, err
}

type Request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type TxJson struct {
	Account         string `json:"Account"`
	Amount          string `json:"Amount"`
	Destination     string `json:"Destination"`
	TransactionType string `json:"TransactionType"`
}

type SubmitParams struct {
	Offline bool   `json:"offline"`
	Secret  string `json:"secret"`
	TxJson  TxJson `json:"tx_json"`
}

func Send(secret, from_addr, to_addr, value string) {
	var req Request
	req.Method = "submit"

	tx_json := new(TxJson)
	tx_json.Account = from_addr
	tx_json.Destination = to_addr
	tx_json.Amount = value
	tx_json.TransactionType = "Payment"

	params := new(SubmitParams)
	params.Offline = false
	params.Secret = "sahx2iJt9dT2BxxANXp4pTsZpHgqr"
	params.TxJson = *tx_json

	req.Params = []SubmitParams{*params}

	body, err := Query(req, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

func main() {
	//generateKeyPairRandom()
	generateKeyPairSeed("sahx2iJt9dT2BxxANXp4pTsZpHgqr")

	QueryBalance("rKM1F8HQoF6LAUapF2n21S38i5LNQu8tHZ")
	QueryBalance("raPuFAdn1xJ5gUvAS1VZQUKBdx6deXLfQK")
	QueryBalance("rSvWdR1UqDG1BThoGT5VyVPzga33yhUwv")
	/*
		txs, err := QueryTx("rKM1F8HQoF6LAUapF2n21S38i5LNQu8tHZ")
		if err != nil {
			panic(err)
		}

		for _, tx := range txs.Transactions {
			fmt.Printf("acc:%s dst:%s value:%s\n", tx.Tx.Account, tx.Tx.Destination, tx.Tx.Amount)
		}
	*/

	/*
		Send(
			"sahx2iJt9dT2BxxANXp4pTsZpHgqr",
			"rKM1F8HQoF6LAUapF2n21S38i5LNQu8tHZ",
			"rSvWdR1UqDG1BThoGT5VyVPzga33yhUwv",
			"80000000",
		)
	*/
}
