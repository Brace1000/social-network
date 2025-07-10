package services

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a given password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // 14 is a good cost
	return string(bytes), err
}

// CheckPasswordHash compares a plain password with a hashed password.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
