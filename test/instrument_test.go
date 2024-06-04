package test

import (
	"testing"

	data "github.com/samjtro/go-schwab-traderapi/market-data"
)

func TestInstrument(t *testing.T) {
	_, err := data.SearchInstrumentFundamental("AAPL")

	if err != nil {
		t.Fatalf(err.Error())
	}
}