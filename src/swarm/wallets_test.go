package swarm

import (
	"fmt"
	"testing"
)

func TestAddingWallets(t *testing.T) {

	b := new(Blockchain)
	b.WalletMapping = make(map[string]uint64)

	//testing adding wallets to the wallet mapping
	b.AddWallet("1", 5)
	b.AddWallet("2", 10)
	b.AddWallet("3", 15)
	b.AddWallet("4", 20)
	b.AddWallet("5", 25)

	//test if wallet is already in the map
	err := b.AddWallet("3", 15)
	if err == nil {
		t.Fatal("Wallet already in map - Failed")
	}

	//test adding wallet with balance of 0
	err = b.AddWallet("6", 0)
	if err == nil {
		t.Fatal("Wallet adding with bal of 0 - Failed")
	}

	//testing if everything is there
	fmt.Println(b.WalletMapping)

	//Testing MoveBal
	b.MoveBal("1", "2", 5)
	b.MoveBal("5", "2", 15)

	//Test Dest DNE
	err = b.MoveBal("2", "6", 5)
	if err == nil {
		t.Fatal("Test Dest DNE - Failed")
	}

	//Test src DNE
	err = b.MoveBal("6", "2", 5)
	if err == nil {
		t.Fatal("Test src DNE - Failed")
	}

	//src wallet does not have enough for transaction
	err = b.MoveBal("5", "1", 15)
	if err == nil {
		t.Fatal("Source wallet does not have enough for transaction - Failed")
	}

	b.MoveBal("5", "1", 10)
	err = b.AddWallet("5", 10)
	if err == nil {
		t.Fatal("Removing 0 balance - Failed")
	}
}
