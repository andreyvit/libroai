package jwt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
	"unsafe"
)

var (
	ErrCorrupted          = errors.New("token corrupted")
	ErrAlg                = errors.New("token uses a wrong algorithm")
	ErrExpired            = errors.New("token expired")
	ErrNotYetValid        = errors.New("token not valid yet")
	ErrTooLong            = errors.New("token too long")
	ErrSignature          = errors.New("token signature invalid")
	ErrSignatureCorrupted = errors.New("token signature corrupted")

	MaxTokenLen        = 8000 // MaxTokenLen is the safety limit to avoid decoding very long data
	ExpectedClaimCount = 10   // ExpectedClaimCount is a starting size for the claims map
)

const (
	TokenID     = "jti" // TokenID is a unique identifier for this token.
	Issuer      = "iss" // Issuer is the principal that issued the token
	Audience    = "aud" // Audience identifies the recipents the token is intended for
	Subject     = "sub" // Subject is the user/account /etc that this token authorizes access to
	IssuedAt    = "iat" // IssuedAt is a Unix timestamp for when the token was issued
	ExpiresAt   = "exp" // ExpiresAt is a Unix timestamp for when the token should expire
	NotBeforeAt = "nbf" // NotBeforeAt is a timestamp this token should not be accepted until

	Forever time.Duration = 1<<63 - 1 // Forever is validity duration of tokens that do not expire

	stackClaimsSpace     = 512
	hs256SignatureEncLen = 43                                     // RawURLEncoding.EncodedLen(sha256.Size)
	hs256Header          = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" // {"alg":"HS256","typ":"JWT"}
)

type Claims map[string]any

func New(subject string, validity time.Duration) Claims {
	return NewAt(subject, validity, time.Now())
}

func NewAt(subject string, validity time.Duration, now time.Time) Claims {
	if validity == 0 {
		// accepting 0 would allow a misconfiguration to escalate into a security issue
		panic("zero validity is invalid, use Forever for non-expiring tokens")
	}

	c := make(Claims, ExpectedClaimCount)
	c[IssuedAt] = now.Unix()
	if validity != Forever {
		c[ExpiresAt] = now.Add(validity).Unix()
	}
	if subject != "" {
		c[Subject] = subject
	}
	return c
}

func (c Claims) String(key string) string {
	if v, ok := c[key].(string); ok {
		return v
	} else {
		return ""
	}
}

func (c Claims) Int64(key string) (int64, bool) {
	switch v := c[key].(type) {
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return n, true
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

func (c Claims) Time(key string) time.Time {
	if v, ok := c.Int64(key); ok && v != 0 {
		return time.Unix(v, 0)
	} else {
		return time.Time{}
	}
}

func (c Claims) ExpiresAt() time.Time {
	return c.Time(ExpiresAt)
}

func (c Claims) Issuer() string {
	return c.String(Issuer)
}

func (c Claims) Subject() string {
	return c.String(Subject)
}

func (c Claims) ValidateTime(tolerance time.Duration) error {
	return c.ValidateTimeAt(tolerance, time.Now())
}

func (c Claims) ValidateTimeAt(tolerance time.Duration, now time.Time) error {
	if exp := c.ExpiresAt(); !exp.IsZero() {
		if now.After(exp.Add(tolerance)) {
			return ErrExpired
		}
	}
	if exp := c.Time(NotBeforeAt); !exp.IsZero() {
		if now.Before(exp.Add(-tolerance)) {
			return ErrNotYetValid
		}
	}
	return nil
}

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func SignHS256String(c Claims, key []byte) string {
	b := SignHS256(c, key, nil)
	return unsafe.String(&b[0], len(b))
}

// SignHS256 produces a signed JWT token from the given claims.
func SignHS256(c Claims, key []byte, buf []byte) []byte {
	claims, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return SignHS256Raw(claims, key, buf)
}

// SignHS256Raw produces a signed JWT token from the given raw claims.
func SignHS256Raw(claims []byte, key []byte, buf []byte) []byte {
	prefixLen := len(hs256Header)
	claimLen := base64.RawURLEncoding.EncodedLen(len(claims))
	tokenLen := prefixLen + 1 + claimLen + 1 + hs256SignatureEncLen

	if len(buf) < tokenLen {
		buf = make([]byte, tokenLen)
	}

	copy(buf, hs256Header)
	buf[prefixLen] = '.'
	base64.RawURLEncoding.Encode(buf[prefixLen+1:], claims)

	var hash [sha256.Size]byte
	alg := hmac.New(sha256.New, key)
	alg.Write(buf[:prefixLen+1+claimLen])
	alg.Sum(hash[:0])

	buf[prefixLen+1+claimLen] = '.'
	base64.RawURLEncoding.Encode(buf[prefixLen+1+claimLen+1:], hash[:])
	return buf
}

// DecodeHS256String verifies JWT token signature and decodes its claims.
func DecodeHS256String(token string, key []byte) (Claims, error) {
	return DecodeHS256(unsafe.Slice(unsafe.StringData(token), len(token)), key)
}

// DecodeHS256 verifies JWT token signature and decodes its claims.
func DecodeHS256(token []byte, key []byte) (Claims, error) {
	if len(token) > MaxTokenLen {
		return nil, ErrTooLong
	}

	i1 := bytes.IndexByte(token, '.')
	if i1 < 0 {
		return nil, ErrCorrupted
	}
	i2 := bytes.IndexByte(token[i1+1:], '.')
	if i2 < 0 {
		return nil, ErrCorrupted
	}
	i2 += i1 + 1

	h := token[:i1]
	if string(h) != hs256Header {
		var hdr header
		err := json.Unmarshal(h, &hdr)
		if err != nil {
			return nil, ErrCorrupted
		}
		if hdr.Typ != "JWT" {
			return nil, ErrCorrupted
		}
		if hdr.Alg != "HS256" {
			return nil, ErrAlg
		}
	}

	{
		var actualHash, expectedHash [sha256.Size]byte
		raw := token[i2+1:]
		if base64.RawURLEncoding.DecodedLen(len(raw)) != len(actualHash) {
			// log.Printf("base64.RawURLEncoding.DecodedLen(len(raw)) %d != len(actualHash) %d", base64.RawURLEncoding.DecodedLen(len(raw)), len(actualHash))
			return nil, ErrSignatureCorrupted
		}
		n, err := base64.RawURLEncoding.Decode(actualHash[:], raw)
		if err != nil || n != len(actualHash) {
			return nil, ErrSignatureCorrupted
		}

		alg := hmac.New(sha256.New, key)
		alg.Write(token[:i2])
		alg.Sum(expectedHash[:0])

		// log.Printf("StringToSign = %q", token[:i2])
		// log.Printf("expectedHash = %q", base64.RawURLEncoding.EncodeToString(expectedHash[:]))
		// log.Printf("actualHash = %q", base64.RawURLEncoding.EncodeToString(actualHash[:]))

		if 1 != subtle.ConstantTimeCompare(actualHash[:], expectedHash[:]) {
			return nil, ErrSignature
		}
	}

	c := make(Claims, ExpectedClaimCount)
	{
		raw := token[i1+1 : i2]
		n := base64.RawURLEncoding.DecodedLen(len(raw))

		// if claims data is small enough, decode into a stack buffer to avoid allocation
		var stackBuf [stackClaimsSpace]byte
		var buf []byte
		if n < cap(stackBuf) {
			buf = stackBuf[:]
		} else {
			buf = make([]byte, n)
		}

		// log.Printf("RawToken = %q", raw)

		n, err := base64.RawURLEncoding.Decode(buf, raw)
		if err != nil {
			return nil, ErrCorrupted
		}

		// log.Printf("JSONToken = %s", buf[:n])

		dec := json.NewDecoder(bytes.NewReader(buf[:n]))
		dec.UseNumber()
		err = dec.Decode(&c)
		if err != nil {
			// log.Printf("JSON err: %v", err)
			return nil, ErrCorrupted
		}
	}

	return c, nil
}
