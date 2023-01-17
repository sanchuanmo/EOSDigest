package tools

import (
	"fmt"
	"log"

	"github.com/polynetwork/eos_relayer/config"
	"github.com/qqtou/eos-go/ecc"
)

type EOSKeyStore struct {
	Ks          *ecc.PrivateKey
	AccountName string
}

func NewEOSKeyStore(sigConfig *config.EOSConfig) []*EOSKeyStore {

	if len(sigConfig.StoreAccounts) == 0 {
		log.Fatal("relayer has no account")
		panic(fmt.Errorf("relayer has no account"))
	}

	EOSAccounts := make([]*EOSKeyStore, len(sigConfig.StoreAccounts))

	for i, account := range sigConfig.StoreAccounts {

		a := &EOSKeyStore{}
		a.Ks, _ = ecc.NewPrivateKey(account["privateKey"])
		a.AccountName = account["accountName"]

		EOSAccounts[i] = a
	}

	return EOSAccounts
}
