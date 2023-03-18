package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
