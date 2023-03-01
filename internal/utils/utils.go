package utils

import (
	"bridge/internal/rpc_error"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

const BcryptCost = bcrypt.DefaultCost

const charset = "abcdefghjklmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789"

var (
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	strReg     = regexp.MustCompile("[^a-z0-9]+")
)

// HashString hashes the provided string returning the hashed value.
func HashString(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CompareHash validates that a hash and value match.
func CompareHash(hash, value string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value)) == nil
}

// ParseDBError parses db errors to return more information to the caller.
func ParseDBError(err error) error {
	if v, ok := err.(*pq.Error); ok {
		switch v.Code {
		case "23505":
			return rpc_error.NewError(codes.AlreadyExists, v.Detail)
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return rpc_error.ErrResourceNotFound
	}

	return rpc_error.ErrServerError
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func Slugify(s string) string {
	return strings.Trim(strReg.ReplaceAllString(strings.ToLower(s), "-"), "-")
}
