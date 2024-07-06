package test

import (
	"fmt"
	"ledgerproto"
)

func main() {
	fmt.Println("Starting request processing...")

	// Example struct initialization that the script will find and wrap
	rsp, err := ledgerproto.CalculateBalanceRequest{
		BalanceName:	ledgerproto.BalanceNameInterestPayable,
		AccountId:	"123456789",
		LegalEntity:	"Monzo Bank Limited",
		Currency:	"GBP",
	}.Send("context").DecodeResponse()

	if err != nil {
		fmt.Println("Error processing request:", err)
		return
	}

	fmt.Println("Response:", rsp)
}
