package storage

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v3"

	"k8s.io/klog"
)

type badgerStorage struct {
	db *badger.DB
}

func NewBadgerStorage(path string) Storage {
	opts := badger.DefaultOptions(path).
		WithNumMemtables(1).
		WithNumLevelZeroTables(1).
		WithValueLogFileSize(1 << 25).
		WithMemTableSize(4 << 20)

	db, err := badger.Open(opts)
	if err != nil {
		klog.Fatal(err)
	}
	bs := &badgerStorage{
		db: db,
	}

	return bs
}

func (fs *badgerStorage) StoreOne(key string, data []byte) error {
	klog.V(8).Infof("storage one key=%s, cache=%s", key, string(data))

	err := fs.update(fs.oneKey(key), data)
	if err != nil {
		klog.Errorf("write one cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (fs *badgerStorage) StoreList(key string, data []byte) error {
	klog.V(8).Infof("storage list key=%s, cache=%s", key, string(data))

	err := fs.update(fs.listKey(key), data)
	if err != nil {
		klog.Errorf("write list cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (fs *badgerStorage) LoadOne(key string) ([]byte, error) {
	data, err := fs.get(fs.oneKey(key))
	if err != nil {
		klog.Errorf("read one cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load one key=%s, cache=%s", key, string(data))
	return data, nil
}

func (fs *badgerStorage) LoadList(key string) ([]byte, error) {
	data, err := fs.get(fs.listKey(key))
	if err != nil {
		klog.Errorf("read list cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load list key=%s, cache=%s", key, string(data))
	return data, nil
}

func (fs *badgerStorage) Delete(key string) error {
	return nil
}

func (fs *badgerStorage) oneKey(key string) string {
	return strings.ReplaceAll(key, "/", "_")
}

func (fs *badgerStorage) listKey(key string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", key, "list"), "/", "_")
}

func (fs *badgerStorage) get(key string) ([]byte, error) {
	var data []byte
	err := fs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fs.oneKey(key)))
		if err != nil {
			klog.Errorf("get one cache %s error: %v", key, err)
			return err
		}

		data, err = item.ValueCopy(nil)
		if err != nil {
			klog.Errorf("copy one cache %s error: %v", key, err)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (fs *badgerStorage) update(key string, data []byte) error {
	err := fs.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), data)
		return err
	})
	return err
}
