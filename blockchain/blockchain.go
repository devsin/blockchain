package blockchain

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

const (
	dbPath = "./tmp/blocks"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockchain() *Blockchain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.Panic(err)
			}
			err = item.Value(func(val []byte) error {
				lastHash = val
				return nil
			})
			return err
		}
	})

	if err != nil {
		log.Panic(err)
	}
	return &Blockchain{lastHash, db}
}

func (chain *Blockchain) AddBlock(data string) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		return err
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := CreateBlock(data, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	if err != nil {
		log.Panic(err)
	}
}

func (chain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{chain.LastHash, chain.Database}
	return iter
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		var encodeBlock []byte
		_ = item.Value(func(val []byte) error {
			encodeBlock = val
			return nil
		})
		block = Deserialize(encodeBlock)
		return err
	})
	if err != nil {
		log.Panic(err)
	}
	iter.CurrentHash = block.PrevHash
	return block
}
