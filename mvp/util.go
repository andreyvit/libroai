package mvp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"mime"
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

func RandomAlpha(n int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 32 characters
	raw := RandomBytes(n)
	for i, b := range raw {
		raw[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(raw)
}

func RandomDigits(n int) string {
	const alphabet = "0123456789"
	raw := RandomBytes(n)
	for i, b := range raw {
		raw[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(raw)
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

func DisableCaching(w http.ResponseWriter) {
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 UTC")
	w.Header().Set("Cache-Control", "no-cache, no-store, no-transform, must-revalidate, private, max-age=0")
	w.Header().Set("Pragma", "no-cache")
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

func DetermineMIMEType(r *http.Request) string {
	s := r.Header.Get("Content-Type")
	if s == "" {
		return ""
	}
	ctype, _, err := mime.ParseMediaType(s)
	if ctype == "" || err != nil {
		return ""
	}
	return ctype
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

// CanonicalEmail returns an email suitable for unique checks.
func CanonicalEmail(email string) string {
	email = strings.TrimSpace(email)
	return strings.ToLower(email)
}

// EmailRateLimitingKey is a slightly paranoid function that maps emails into string keys to use for rate limiting.
func EmailRateLimitingKey(email string) string {
	username, host, found := strings.Cut(email, "@")
	if !found {
		return "invalid" // use single key for all invalid emails to cut down on stupid shenanigans
	}
	// get rid of local part (after +)
	username, _, _ = strings.Cut(username, "+")
	username = strings.Map(keepOnlyLettersAndNumbers, username)
	email = username + "@" + host
	email = strings.ToLower(email)
	return email
}

var lettersAndNumbers = [128]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0}

func keepOnlyLettersAndNumbers(r rune) rune {
	if r < 128 && lettersAndNumbers[r] != 0 {
		return r
	} else {
		return -1
	}
}

func isWhitespaceOrComma(r rune) bool {
	return r == ' ' || r == ','
}
