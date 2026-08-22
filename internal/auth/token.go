// 轻量级 JWT（JSON Web Token）认证管理器，
// 使用 HMAC-SHA256（HS256） 对 Token 进行签名和验证
// 没有使用第三方 JWT 库，而是直接通过 Go 标准库的 HMAC + SHA256
// + Base64 + JSON 实现了 HS256 JWT 的核心机制
// 登录成功 -> Issue(userID, role) -> 生成 JWT
// -> 客户端携带 JWT 请求 API -> Parse(token)
// -> 验证签名 + 过期时间 + 用户ID + 角色 -> 得到 Claims
// -> 知道“谁在访问、是什么角色”

// JWT 主要用于身份认证：让后端知道“这个请求是谁发的，以及他有什么权限”
// 实际一般直接使用成熟的 JWT 库, 如/golang-jwt/jwt/

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

// Token 里保存的信息, 对应 JWT Payload
// sub：用户 ID
// role：用户角色
// iat：Token 签发时间
// exp：Token 过期时间
// json:"sub,string" 表示把 int64 的 123 编码成字符串 "123"
type Claims struct {
	UserID int64  `json:"sub,string"`
	Role   string `json:"role"`
	Issued int64  `json:"iat"`
	Expiry int64  `json:"exp"`
}

// 生成和验证 Token
// secret：JWT 签名密钥
// ttl：Token 有效期，例如 2 * time.Hour
// now：获取当前时间，默认 time.Now
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// 创建 JWT 管理器
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

// 签发 JWT
func (m *Manager) Issue(userID int64, role string) (string, error) {
	now := m.now().UTC()
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(Claims{UserID: userID, Role: role, Issued: now.Unix(), Expiry: now.Add(m.ttl).Unix()})
	if err != nil {
		return "", fmt.Errorf("encode token claims: %w", err)
	}

	// 最终生成标准 JWT 结构: Header.Payload.Signature
	unsigned := encode(header) + "." + encode(payload)
	return unsigned + "." + encode(m.sign(unsigned)), nil
}

// 验证并解析 JWT
func (m *Manager) Parse(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("malformed token")
	}

	// 重新计算签名, 和 Token 自带的签名比较
	// 使用 hmac.Equal 而不是普通的 ==，可以避免一些时序攻击（timing attack）
	if !hmac.Equal(m.sign(parts[0]+"."+parts[1]), decodeOrNil(parts[2])) {
		return Claims{}, errors.New("invalid token signature")
	}

	// 解码 Payload
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

// 生成签名
func (m *Manager) sign(value string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

// 使用 JWT 所需的 Base64 URL 编码
func encode(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }
func decodeOrNil(value string) []byte {
	decoded, _ := base64.RawURLEncoding.DecodeString(value)
	return decoded
}
