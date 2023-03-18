package deployment

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
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

func ensureSkippingOSExists(err error) {
	if os.IsExist(err) {
		return
	}
	if err != nil {
		panic(err)
	}
}

func atoi(s string) int {
	return must(strconv.Atoi(s))
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

func shellQuoteCmdline(name string, args ...string) string {
	var buf strings.Builder
	buf.WriteString(shellQuote(name))
	for _, arg := range args {
		buf.WriteByte(' ')
		buf.WriteString(shellQuote(arg))
	}
	return buf.String()
}

func shellQuote(source string) string {
	const specialChars = "\\'\"`${[|&;<>()*?! \t\n~"
	const specialInDouble = "$\\\""

	var buf strings.Builder
	buf.Grow(len(source) + 10)

	// pick quotation style, preferring single quotes
	if !strings.ContainsAny(source, specialChars) && len(source) > 0 {
		buf.WriteString(source)
	} else if !strings.ContainsRune(source, '\'') {
		buf.WriteByte('\'')
		buf.WriteString(source)
		buf.WriteByte('\'')
	} else {
		buf.WriteByte('"')
		for {
			i := strings.IndexAny(source, specialInDouble)
			if i < 0 {
				break
			}
			buf.WriteString(source[:i])
			buf.WriteByte('\\')
			buf.WriteByte(source[i])
			source = source[i+1:]
		}
		buf.WriteString(source)
		buf.WriteByte('"')
	}
	return buf.String()
}
