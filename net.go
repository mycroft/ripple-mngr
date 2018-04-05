package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type RequestResult struct {
	Result interface{} `json:"result"`
}

func Query(r interface{}, out interface{}) ([]byte, error) {
	jsonStr, err := json.Marshal(r)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequest("POST", fApiUrl, bytes.NewBuffer(jsonStr))
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
