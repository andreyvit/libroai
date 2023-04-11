package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/andreyvit/buddyd/internal/flake"
)

type Store struct {
	dataDir string
	idGen   *flake.Gen
}

type collection struct {
	name      string
	singleton bool
	ext       string
}

func setupStore(path string, node int) *Store {
	time.Sleep(2 * time.Millisecond) // ensure unique snowflake IDs
	for _, coll := range collections {
		if !coll.singleton {
			ensure(os.MkdirAll(filepath.Join(path, coll.name), 0755))
		}
	}
	return &Store{
		dataDir: path,
		idGen:   flake.NewGen(0, uint64(node)),
	}
}

func (store *Store) NewID() flake.ID {
	return store.idGen.New()
}

func (store *Store) lookupPath(coll *collection, key string) string {
	if coll.singleton {
		if key != "" {
			panic("non-empty key for singleton collection")
		}
		return filepath.Join(store.dataDir, coll.name+coll.ext)
	} else {
		if key == "" {
			panic("empty key for list collection")
		}
		return filepath.Join(store.dataDir, coll.name, key+coll.ext)
	}
}

func (store *Store) Exists(coll *collection, key string) bool {
	fn := store.lookupPath(coll, key)
	_, err := os.Stat(fn)
	return err == nil
}

func (store *Store) LoadRaw(coll *collection, key string) []byte {
	fn := store.lookupPath(coll, key)
	log.Printf("%s : %s  ==>  %s", coll.name, key, fn)
	b, err := os.ReadFile(fn)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	ensure(err)
	return b
}

func (store *Store) Load(coll *collection, key string, ptr any) bool {
	raw := store.LoadRaw(coll, key)
	if raw == nil {
		return false
	}
	decode(raw, ptr)
	return true
}

func (store *Store) Save(coll *collection, key string, data any) {
	raw := encode(data)
	fn := store.lookupPath(coll, key)
	ensure(writeFileAtomic(fn, raw, 0644))
}

func encode(v any) []byte {
	switch v := v.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return must(json.MarshalIndent(v, "", "  "))
	}
}

func decode(b []byte, ptr any) {
	switch ptr := ptr.(type) {
	case *[]byte:
		*ptr = b
	case *string:
		*ptr = string(b)
	default:
		ensure(json.Unmarshal(b, ptr))
	}
}
