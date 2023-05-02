package mvp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"unsafe"
)

func As[T, Base any](base *Base) *T {
	return (*T)(unsafe.Pointer(base))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func ensure(err error) {
	if err != nil {
		panic(err)
	}
}

func RandomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(fmt.Errorf("cannot read %d random bytes: %w", len, err))
	}
	return b
}

func RandomHex(len int) string {
	b := RandomBytes((len + 1) / 2)
	return hex.EncodeToString(b)[:len]
}

func WriteFileAtomic(path string, data []byte, perm fs.FileMode) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), ".~"+filepath.Base(path)+".*")
	if err != nil {
		return err
	}

	var ok, closed bool
	defer func() {
		if !closed {
			temp.Close()
		}
		if !ok {
			os.Remove(temp.Name())
		}
	}()

	err = temp.Chmod(perm)
	if err != nil {
		return err
	}

	_, err = temp.Write(data)
	if err != nil {
		return err
	}

	err = temp.Close()
	closed = true
	if err != nil {
		return err
	}

	err = os.Rename(temp.Name(), path)
	if err != nil {
		return err
	}

	ok = true
	return nil
}

// TrimPort removes :port part (if any) from the given string and returns just the hostname.
func TrimPort(host string) string {
	h, _, _ := net.SplitHostPort(host)
	if h == "" {
		return host
	}
	return h
}

func MarkPublicImmutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
}

func MarkPublicMutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, no-cache, max-age=0")
}

func MarkPrivateMutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-cache, max-age=0")
}

type action func()

func (_ action) String() string {
	return ""
}

func (_ action) IsBoolFlag() bool {
	return true
}

func (f action) Set(string) error {
	f()
	os.Exit(0)
	return nil
}
