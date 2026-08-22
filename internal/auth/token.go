package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	UserID int64  `json:"sub,string"`
	Role   string `json:"role"`
	Issued int64  `json:"iat"`
	Expiry int64  `json:"exp"`
}

type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (m *Manager) Issue(userID int64, role string) (string, error) {
	now := m.now().UTC()
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(Claims{UserID: userID, Role: role, Issued: now.Unix(), Expiry: now.Add(m.ttl).Unix()})
	if err != nil {
		return "", fmt.Errorf("encode token claims: %w", err)
	}
	unsigned := encode(header) + "." + encode(payload)
	return unsigned + "." + encode(m.sign(unsigned)), nil
}

func (m *Manager) Parse(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("malformed token")
	}
	if !hmac.Equal(m.sign(parts[0]+"."+parts[1]), decodeOrNil(parts[2])) {
		return Claims{}, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, errors.New("invalid token payload")
	}
	var wire struct {
		Sub    json.RawMessage `json:"sub"`
		Role   string          `json:"role"`
		Issued int64           `json:"iat"`
		Expiry int64           `json:"exp"`
	}
	if err := json.Unmarshal(payload, &wire); err != nil {
		return Claims{}, errors.New("invalid token claims")
	}
	var sub string
	if err := json.Unmarshal(wire.Sub, &sub); err != nil {
		return Claims{}, errors.New("invalid token subject")
	}
	userID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil || userID <= 0 || wire.Expiry <= m.now().UTC().Unix() || (wire.Role != "customer" && wire.Role != "admin") {
		return Claims{}, errors.New("invalid or expired token")
	}
	return Claims{UserID: userID, Role: wire.Role, Issued: wire.Issued, Expiry: wire.Expiry}, nil
}

func (m *Manager) sign(value string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
func encode(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }
func decodeOrNil(value string) []byte {
	decoded, _ := base64.RawURLEncoding.DecodeString(value)
	return decoded
}
