package db

import (
	"encoding/base64"
	"encoding/json"

	"github.com/renproject/swapperd/core/wallet/transfer"
	"github.com/renproject/swapperd/foundation/blockchain"
	"github.com/renproject/swapperd/foundation/swap"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	TableSwaps      = [8]byte{}
	TableSwapsStart = [40]byte{}
	TableSwapsLimit = [40]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	TablePendingSwaps      = [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	TablePendingSwapsStart = [40]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	TablePendingSwapsLimit = [40]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	TableSwapReceipts      = [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
	TableSwapReceiptsStart = [40]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	TableSwapReceiptsLimit = [40]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

type Storage interface {
	PutSwap(blob swap.SwapBlob) error
	DeletePendingSwap(swapID swap.SwapID) error
	PendingSwaps() ([]swap.SwapBlob, error)

	PutTransfer(transfer transfer.TransferReceipt) error
	Transfers() ([]transfer.TransferReceipt, error)
	UpdateTransferReceipt(updateReceipt transfer.UpdateReceipt) error

	PendingSwap(swapID swap.SwapID) (swap.SwapBlob, error)
	PutReceipt(receipt swap.SwapReceipt) error
	UpdateReceipt(receiptUpdate swap.ReceiptUpdate) error
	Receipts() ([]swap.SwapReceipt, error)
	Receipt(swapID swap.SwapID) (swap.SwapReceipt, error)
	LoadCosts(swapID swap.SwapID) (blockchain.Cost, blockchain.Cost)
}

type dbStorage struct {
	db *leveldb.DB
}

func New(db *leveldb.DB) Storage {
	return &dbStorage{
		db: db,
	}
}

func (db *dbStorage) PutSwap(blob swap.SwapBlob) error {
	blob.Password = ""
	swapData, err := json.Marshal(blob)
	if err != nil {
		return err
	}
	id, err := base64.StdEncoding.DecodeString(string(blob.ID))
	if err != nil {
		return err
	}
	if err := db.db.Put(append(TablePendingSwaps[:], id...), swapData, nil); err != nil {
		return err
	}
	return db.db.Put(append(TableSwaps[:], id...), swapData, nil)
}

func (db *dbStorage) DeletePendingSwap(swapID swap.SwapID) error {
	id, err := base64.StdEncoding.DecodeString(string(swapID))
	if err != nil {
		return err
	}
	return db.db.Delete(append(TablePendingSwaps[:], id...), nil)
}

func (db *dbStorage) PendingSwap(swapID swap.SwapID) (swap.SwapBlob, error) {
	id, err := base64.StdEncoding.DecodeString(string(swapID))
	if err != nil {
		return swap.SwapBlob{}, err
	}
	swapBlobBytes, err := db.db.Get(append(TablePendingSwaps[:], id...), nil)
	if err != nil {
		return swap.SwapBlob{}, err
	}
	blob := swap.SwapBlob{}
	if err := json.Unmarshal(swapBlobBytes, &blob); err != nil {
		return swap.SwapBlob{}, err
	}
	return blob, nil
}

func (db *dbStorage) PendingSwaps() ([]swap.SwapBlob, error) {
	iterator := db.db.NewIterator(&util.Range{Start: TablePendingSwapsStart[:], Limit: TablePendingSwapsLimit[:]}, nil)
	defer iterator.Release()
	pendingSwaps := []swap.SwapBlob{}
	for iterator.Next() {
		value := iterator.Value()
		swap := swap.SwapBlob{}
		if err := json.Unmarshal(value, &swap); err != nil {
			return pendingSwaps, err
		}
		pendingSwaps = append(pendingSwaps, swap)
	}
	return pendingSwaps, iterator.Error()
}
