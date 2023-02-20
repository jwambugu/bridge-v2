package utils

import "golang.org/x/crypto/bcrypt"

const BcryptCost = bcrypt.DefaultCost

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
