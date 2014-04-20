package quorum

import (
	"fmt"
	"testing"
)

func TestAddingWallets(t *testing.T) {

	s := new(State)
	s.wallets = make(map[string]uint64)

	//testing adding wallets to the wallet mapping
	s.AddWallet("1", 5)
	s.AddWallet("2", 10)
	s.AddWallet("3", 15)
	s.AddWallet("4", 20)
	s.AddWallet("5", 25)

	//test if wallet is already in the map
	err := s.AddWallet("3", 15)
	if err == nil {
		t.Fatal("Wallet already in map - Failed")
	}

	//test adding wallet with balance of 0
	err = s.AddWallet("6", 0)
	if err == nil {
		t.Fatal("Wallet adding with bal of 0 - Failed")
	}

	//testing if everything is there
	fmt.Println(s.wallets)

	//Testing MoveBal
	s.MoveBal("1", "2", 5)
	s.MoveBal("5", "2", 15)

	//Test Dest DNE
	err = s.MoveBal("2", "6", 5)
	if err == nil {
		t.Fatal("Test Dest DNE - Failed")
	}

	//Test src DNE
	err = s.MoveBal("6", "2", 5)
	if err == nil {
		t.Fatal("Test src DNE - Failed")
	}

	//src wallet does not have enough for transaction
	err = s.MoveBal("5", "1", 15)
	if err == nil {
		t.Fatal("Source wallet does not have enough for transaction - Failed")
	}

	s.MoveBal("5", "1", 10)
	err = s.AddWallet("5", 10)
	if err == nil {
		t.Fatal("Removing 0 balance - Failed")
	}
}
