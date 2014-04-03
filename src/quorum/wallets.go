package quorum

import (
	"errors"
)

func (b *Blockchain) AddWallet(id string, bal uint64) (err error) {

	if bal == 0 {
		return errors.New("Cannot add balance of 0!")
	}
	elem, ok := b.WalletMapping[id]
	if ok {
		return errors.New("This wallet already exists!")
	} else {
		elem = bal
		b.WalletMapping[id] = elem
	}
	return nil
}

func (b *Blockchain) MoveBal(src string, dest string, amt uint64) (err error) {

	//check to make sure the wallets exist
	elem, ok := b.WalletMapping[dest]
	if ok {
		return errors.New("Destination wallet does not exist!")
	}
	elem, ok = b.WalletMapping[src]
	if ok {
		return errors.New("Source wallet does not exist!")
	}

	//check balance editting
	tmp := elem - amt
	if tmp < 0 {
		return errors.New("Source wallet does not have enough coins!")
	} else if tmp == 0 {
		delete(b.WalletMapping, src)
	} else {
		b.WalletMapping[src] = tmp
	}

	//change balance in destination
	b.WalletMapping[dest] += tmp

	return nil
}
