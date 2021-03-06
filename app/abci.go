package app

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/libs/kv"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/tendermint/tendermint/abci/types"
)

type KVStoreApplication struct {
	types.Application
	store        store.CommitKVStore
}

func NewKVStoreApplication(db dbm.DB) *KVStoreApplication {
	app := &KVStoreApplication{
		Application: types.NewBaseApplication(),
	}
	app.initDB(db)
	return app
}

func (app *KVStoreApplication) initDB(db dbm.DB) (err error) {
	app.store, err = iavl.LoadStore(db, store.CommitID{}, store.PruneNothing, false)
	if err != nil {
		return err
	}
	return nil
}

func (app *KVStoreApplication) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	var key, value []byte
	parts := bytes.Split(req.Tx, []byte("="))
	if len(parts) == 2 {
		key, value = parts[0], parts[1]
	} else {
		key, value = req.Tx, req.Tx
	}

	app.store.Set(key, value)

	events := []types.Event{
		{
			Type: "set",
			Attributes: []kv.Pair{
				{Key: []byte("key"), Value: key},
				{Key: []byte("hash"), Value: value},
			},
		},
	}
	return types.ResponseDeliverTx{Code: code.CodeTypeOK, Events: events}
}

func (app *KVStoreApplication) Commit() types.ResponseCommit {
	commitID := app.store.Commit()
	return types.ResponseCommit{ Data: commitID.Hash }
}

func (app *KVStoreApplication) Query(req types.RequestQuery) types.ResponseQuery {
	iavlStore := app.store.(*iavl.Store)

	res := iavlStore.Query(types.RequestQuery{
		Path:  "/key", // required path to get key/value+proof
		Data:  req.Data,
		Prove: true,
	})

	return res
}
