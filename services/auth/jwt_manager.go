package auth

import (
	"bridge/api/v1/pb"
	"bridge/core/config"
	"fmt"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
	"time"
)

type JWTManager interface {
	Generate(user *pb.User, duration time.Duration) (string, error)
	Verify(token string) (*Payload, error)
}

type (
	pasetoToken struct {
		key    []byte
		paseto *paseto.V2
	}

	Payload struct {
		Audience   string
		Expiration time.Time
		IssuedAt   time.Time
		Issuer     string
		NotBefore  time.Time
		Subject    uint64
		User       *pb.User
	}
)

func newPayload(user *pb.User, duration time.Duration) Payload {
	var (
		appName = config.Get[string](config.AppName, "bridge")
		now     = time.Now()
	)

	return Payload{
		Audience:   appName,
		Expiration: time.Now().Add(duration),
		IssuedAt:   now,
		Issuer:     appName,
		NotBefore:  now,
		Subject:    user.ID,
		User:       user,
	}
}

func (p *pasetoToken) Generate(user *pb.User, duration time.Duration) (string, error) {
	payload := newPayload(user, duration)
	return p.paseto.Encrypt(p.key, payload, nil)
}

func (p *pasetoToken) Verify(token string) (*Payload, error) {
	payload := &Payload{}
	if err := p.paseto.Decrypt(token, p.key, &payload, nil); err != nil {
		return nil, err
	}
	return payload, nil
}

// NewPasetoToken create a new paseto token
func NewPasetoToken(key string) (JWTManager, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	p := &pasetoToken{
		key:    []byte(key),
		paseto: paseto.NewV2(),
	}

	return p, nil
}
