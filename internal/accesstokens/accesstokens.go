package accesstokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type Configuration struct {
	// Keys are a set of keys to sign tokens. The first one is used for new tokens.
	// Other keys are accepted when validating tokens, to allow key rotation.
	Keys [][]byte

	// Prefixes are added in front of the tokens to help identify them.
	// The first one is used for new tokens. Others are accepted when
	// validating tokens to allow prefix changes.
	Prefixes []string

	Validity time.Duration
}

const (
	TimeFormat = "20060102150405"

	sep = "-"

	v1 = "1"

	Infinite time.Duration = time.Hour * 24 * 365 * 100
)

type components struct {
	prefixIndex int
	ver         string
	timeStr     string
	account     string
}

type Token struct {
	Account    string
	Creation   time.Time
	Expiration time.Time
	Upgradable bool
}

func (t Token) DebugString() string {
	if t == (Token{}) {
		return "<none>"
	}
	var b strings.Builder
	b.WriteString(t.Account)
	b.WriteString(" from=")
	b.WriteString(t.Creation.Format(TimeFormat))
	b.WriteString(" till=")
	b.WriteString(t.Expiration.Format(TimeFormat))
	if t.Upgradable {
		b.WriteString(" upgradable")
	}
	return b.String()
}

var (
	Invalid    = errors.New("invalid token")
	InvalidSig = errors.New("invalid token signature")
	Expired    = errors.New("expired token")
)

func (conf *Configuration) SignAt(now time.Time, account string) string {
	if len(conf.Keys) == 0 || len(conf.Prefixes) == 0 || len(conf.Prefixes[0]) == 0 {
		panic("accesstokens: not configured")
	}

	nowStr := now.UTC().Format(TimeFormat)
	msg := conf.computeMessage(nowStr, account)
	auth := hmacSHA256(conf.Keys[0], []byte(msg))
	return msg + sep + auth
}

func (conf *Configuration) ValidateAt(now time.Time, token string) (Token, error) {
	i := strings.LastIndex(token, sep)
	if i < 0 {
		return Token{}, Invalid
	}
	auth, msg := token[i+1:], token[:i]

	c, err := conf.parseMessage(msg)
	if err != nil {
		return Token{}, err
	}

	keyIndex := -1
	for i, key := range conf.Keys {
		expected := hmacSHA256(key, []byte(msg))
		if subtle.ConstantTimeCompare([]byte(auth), []byte(expected)) == 1 {
			keyIndex = i
			break
		}
	}
	if keyIndex < 0 {
		return Token{}, InvalidSig
	}

	creation, err := time.ParseInLocation(TimeFormat, c.timeStr, time.UTC)
	if err != nil {
		return Token{}, Invalid
	}

	expiration := creation.Add(conf.Validity)
	if now.After(expiration) {
		return Token{}, Expired
	}

	return Token{
		Account:    c.account,
		Creation:   creation,
		Expiration: expiration,
		Upgradable: (c.prefixIndex > 0) || (keyIndex > 0),
	}, nil
}

func (conf *Configuration) computeMessage(nowStr, account string) string {
	return conf.Prefixes[0] + v1 + sep + nowStr + sep + account
}

func (conf *Configuration) parseMessage(msg string) (components, error) {
	c := components{prefixIndex: -1}
	for i, p := range conf.Prefixes {
		if strings.HasPrefix(msg, p) {
			msg = msg[len(p):]
			c.prefixIndex = i
			break
		}
	}
	if c.prefixIndex < 0 {
		return c, Invalid
	}

	i := strings.Index(msg, sep)
	if i < 0 {
		return c, Invalid
	}
	c.ver, msg = msg[:i], msg[i+1:]

	if c.ver != v1 {
		return c, Invalid
	}

	i = strings.Index(msg, sep)
	if i < 0 {
		return c, Invalid
	}
	c.timeStr, c.account = msg[:i], msg[i+1:]

	return c, nil
}

func hmacSHA256(key, message []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

func ParseKeys(s string) ([][]byte, error) {
	var keys [][]byte
	for _, ks := range strings.FieldsFunc(s, isWhitespaceOrComma) {
		if ks == "" {
			continue
		}
		key, err := hex.DecodeString(ks)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func isWhitespaceOrComma(r rune) bool {
	return r == ' ' || r == ','
}

type Keys [][]byte

func KeysVar(v *[][]byte) *Keys {
	return (*Keys)(v)
}

func (v Keys) String() string {
	var buf strings.Builder
	for i, k := range v {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(hex.EncodeToString(k))
	}
	return buf.String()
}

func (v Keys) Get() interface{} {
	return [][]byte(v)
}

func (v *Keys) Set(raw string) (err error) {
	*v, err = ParseKeys(raw)
	return
}
