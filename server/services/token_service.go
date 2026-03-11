package services

import (
	"crypto/rand"
	"encoding/hex"
)

type TokenService struct{}

func NewTokenService() *TokenService { return &TokenService{} }

func (s *TokenService) Generate(prefix string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}