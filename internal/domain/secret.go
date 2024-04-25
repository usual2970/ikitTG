package domain

import (
	"context"
)

type Secret struct {
	Meta
	Uri         string            `json:"uri"`
	ApiKey      string            `json:"apiKey"`
	SecretKey   string            `json:"secretKey"`
	Description string            `json:"description"`
	Ext         map[string]string `json:"ext"`
}

type ISecretRepository interface {
	Get(ctx context.Context, filter string) (*Secret, error)
}
