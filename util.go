package main

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
	"strings"
)

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

func loadEnv(fn string) {
	raw := string(must(os.ReadFile(fn)))
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok || strings.ContainsAny(k, " ") {
			continue
		}
		os.Setenv(strings.TrimSpace(k), strings.TrimSpace(v))
	}
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(fmt.Errorf("cannot read %d random bytes: %w", len, err))
	}
	return b
}

func randomHex(len int) string {
	b := randomBytes((len + 1) / 2)
	return hex.EncodeToString(b)[:len]
}

func markPublicImmutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
}

func markPublicMutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, no-cache, max-age=0")
}

func markPrivateMutable(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-cache, max-age=0")
}

// trimPort removes :port part (if any) from the given string and returns just the hostname.
func trimPort(host string) string {
	h, _, _ := net.SplitHostPort(host)
	if h == "" {
		return host
	}
	return h
}

func writeFileAtomic(path string, data []byte, perm fs.FileMode) (err error) {
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
