package backend

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgraph-io/badger/v3"
)

type Message struct {
	Key  string
	Exp  time.Time
	Data []byte
}

var letterRunes = []rune("23456789abcdefgjkmpqrstuvwxyzABCDEFGJKMPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func (backend *Backend) makeMessage(duration time.Duration, msg string, pin string) (*Message, error) {
	message := &Message{
		Key: RandStringRunes(30),
		Exp: time.Now().Add(duration),
	}

	if err := backend.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("msg/%s", message.Key))

		req := CryptRequest{
			ID:   message.Key,
			Data: []byte(msg),
			Pin:  pin,
		}

		data, err := backend.crypt.Encrypt(req)
		if err != nil {
			return err
		}

		e := badger.NewEntry(key, data).WithTTL(duration)

		return txn.SetEntry(e)
	}); err != nil {
		return nil, err
	}

	return message, nil
}

func (backend *Backend) loadMessage(key string, pin string) (*Message, error) {
	message := &Message{}
	var data []byte

	dbKey := []byte(fmt.Sprintf("msg/%s", key))

	if err := backend.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(dbKey)
		if err != nil {
			return err
		}

		if item.IsDeletedOrExpired() {
			return badger.ErrKeyNotFound
		}

		item.Value(func(val []byte) error {
			data = val
			return nil
		})

		return nil
	}); err != nil {
		return nil, err
	}

	var err error

	req := CryptRequest{
		ID:   key,
		Data: data,
		Pin:  pin,
	}

	message.Data, err = backend.crypt.Decrypt(req)
	if err != nil {
		return nil, err
	}

	if err := backend.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(dbKey)
	}); err != nil {
		return nil, err
	}

	return message, nil
}
