// Package auth 提供商户令牌的签发与校验。
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/chihqiang/ten-go/rest/token"
)

// 令牌类型声明，签发与校验时用于区分访问令牌和刷新令牌。
const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"

	claimSubject   = "sub"
	claimIssuedAt  = "iat"
	claimExpiresAt = "exp"
	claimIssuer    = "iss"

	// minSecretLength 签名密钥最小长度（沿用旧版 token 包的校验）。
	minSecretLength = 8
)

// Config 商户令牌配置。
type Config struct {
	// Secret 签名密钥（敏感信息走环境变量）。
	Secret string `json:"secret,optional"`
	// Issuer 签发者。
	Issuer string `json:"issuer,optional"`
	// AccessTokenExpire 访问令牌有效期。
	AccessTokenExpire time.Duration `json:"access_token_expire,default=2h"`
	// RefreshTokenExpire 刷新令牌有效期。
	RefreshTokenExpire time.Duration `json:"refresh_token_expire,default=168h"`
}

// TokenPair 商户令牌对。
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Claims 是校验通过后的令牌载荷。
type Claims struct {
	Subject   string
	Issuer    string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Data      map[string]any
}

// Tokens 商户令牌签发与校验。
// 访问令牌与刷新令牌使用由主密钥派生的不同签名密钥，因此二者不能互相冒用。
type Tokens struct {
	access        *token.Parser
	refresh       *token.Parser
	issuer        string
	accessExpire  time.Duration
	refreshExpire time.Duration
}

// NewTokens 按配置创建访问/刷新令牌解析器。
func NewTokens(c Config) (*Tokens, error) {
	if c.Secret == "" {
		return nil, errors.New("auth: token secret is empty")
	}
	if len(c.Secret) < minSecretLength {
		return nil, errors.New("auth: token secret must be at least 8 characters")
	}
	accessExpire := c.AccessTokenExpire
	if accessExpire <= 0 {
		accessExpire = 2 * time.Hour
	}
	refreshExpire := c.RefreshTokenExpire
	if refreshExpire <= 0 {
		refreshExpire = 168 * time.Hour
	}
	return &Tokens{
		access:        token.NewParser(token.Config{Secret: deriveSecret(c.Secret, tokenTypeAccess)}),
		refresh:       token.NewParser(token.Config{Secret: deriveSecret(c.Secret, tokenTypeRefresh)}),
		issuer:        c.Issuer,
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
	}, nil
}

// Issue 为商户签发访问令牌与刷新令牌。
func (t *Tokens) Issue(accountID uint, username string) (*TokenPair, error) {
	subject := strconv.FormatUint(uint64(accountID), 10)
	now := time.Now()

	access, err := t.access.Generate(t.claims(subject, now, t.accessExpire, map[string]any{
		"token_type": tokenTypeAccess,
		"username":   username,
		"role":       "account",
	}))
	if err != nil {
		return nil, err
	}
	refresh, err := t.refresh.Generate(t.claims(subject, now, t.refreshExpire, map[string]any{
		"token_type": tokenTypeRefresh,
	}))
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    now.Add(t.accessExpire),
	}, nil
}

// ParseAccess 校验访问令牌并返回 claims。
func (t *Tokens) ParseAccess(raw string) (*Claims, error) {
	claims, err := t.access.Parse(raw)
	if err != nil {
		return nil, err
	}
	verified, err := t.verifiedClaims(claims)
	if err != nil {
		return nil, err
	}
	if verified.Data["token_type"] != tokenTypeAccess {
		return nil, token.ErrInvalidToken
	}
	return verified, nil
}

// ParseRefresh 校验刷新令牌并返回商户 ID。
func (t *Tokens) ParseRefresh(raw string) (uint, error) {
	claims, err := t.refresh.Parse(raw)
	if err != nil {
		return 0, err
	}
	verified, err := t.verifiedClaims(claims)
	if err != nil {
		return 0, err
	}
	if verified.Data["token_type"] != tokenTypeRefresh {
		return 0, token.ErrInvalidToken
	}
	id, err := strconv.ParseUint(verified.Subject, 10, 64)
	if err != nil || id == 0 {
		return 0, token.ErrInvalidToken
	}
	return uint(id), nil
}

// claims 拼装底层 token 包所需的完整载荷（sub/iat/exp/iss + 业务 claim）。
func (t *Tokens) claims(subject string, now time.Time, expire time.Duration, data map[string]any) jwtlib.MapClaims {
	claims := jwtlib.MapClaims{
		claimSubject:   subject,
		claimIssuedAt:  now.Unix(),
		claimExpiresAt: now.Add(expire).Unix(),
	}
	if t.issuer != "" {
		claims[claimIssuer] = t.issuer
	}
	for key, value := range data {
		claims[key] = value
	}
	return claims
}

// verifiedClaims 把 MapClaims 转为业务 Claims，并保留旧的 issuer 校验语义。
func (t *Tokens) verifiedClaims(raw jwtlib.MapClaims) (*Claims, error) {
	claims := &Claims{Data: make(map[string]any, len(raw))}
	for key, value := range raw {
		switch key {
		case claimSubject:
			claims.Subject, _ = value.(string)
		case claimIssuer:
			claims.Issuer, _ = value.(string)
		case claimIssuedAt:
			claims.IssuedAt = unixTime(value)
		case claimExpiresAt:
			claims.ExpiresAt = unixTime(value)
		default:
			claims.Data[key] = value
		}
	}
	if claims.Subject == "" {
		return nil, token.ErrNoClaims
	}
	if t.issuer != "" && claims.Issuer != t.issuer {
		return nil, token.ErrInvalidToken
	}
	return claims, nil
}

func unixTime(value any) time.Time {
	switch typed := value.(type) {
	case float64:
		return time.Unix(int64(typed), 0)
	case int64:
		return time.Unix(typed, 0)
	default:
		return time.Time{}
	}
}

// deriveSecret 由主密钥派生指定用途的签名密钥。
func deriveSecret(secret, scope string) string {
	sum := sha256.Sum256([]byte(secret + ":" + scope))
	return hex.EncodeToString(sum[:])
}
