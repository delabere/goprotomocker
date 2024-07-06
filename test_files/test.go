package test

import (
	"fmt"
	"ledgerproto"
)

func main() {
	fmt.Println("Starting request processing...")

	rsp, err := ledgerproto.CalculateBalanceRequest{
		BalanceName: ledgerproto.BalanceNameInterestPayable,
		AccountId:   "123456789",
		LegalEntity: "Monzo Bank Limited",
		Currency:    "GBP",
	}.Send(ctx).DecodeResponse()

	ledgerproto.CalculateBalanceRequest{
		BalanceName: ledgerproto.BalanceNameInterestPayable,
		AccountId:   "123456789",
		LegalEntity: "Monzo Bank Limited",
		Currency:    "GBP",
	}

	fmt.Println("Response:", rsp)
}
