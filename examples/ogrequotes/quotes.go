package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type quoteLine struct {
	Speaker string `json:"speaker"`
	Line    string `json:"line"`
}

type quote []quoteLine

//go:embed quotes.json
var ogreQuotesFile []byte

func loadOgreQuotes() ([]quote, error) {
	var quotes []quote

	if err := json.Unmarshal(ogreQuotesFile, &quotes); err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON quotes file: %w", err)
	}

	return quotes, nil
}
