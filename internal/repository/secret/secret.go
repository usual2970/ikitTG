package secret

import (
	"context"
	"ikit-api/internal/domain"
	"ikit-api/internal/util/app"
	"sync"
)

var once sync.Once
var instance domain.ISecretRepository

type repository struct{}

func NewRepository() domain.ISecretRepository {
	once.Do(func() {
		instance = &repository{}
	})
	return instance
}

func (r *repository) Get(ctx context.Context, filter string) (*domain.Secret, error) {
	record, err := app.Get().Dao().FindFirstRecordByFilter("secrets",
		filter,
	)
	if err != nil {
		return nil, err
	}

	ext := make(map[string]string)

	record.UnmarshalJSONField("ext", &ext)
	meta := domain.Meta{
		Id:      record.Id,
		Created: record.GetTime("created"),
		Updated: record.GetTime("updated"),
	}
	rs := &domain.Secret{

		Uri:         record.GetString("uri"),
		ApiKey:      record.GetString("api_key"),
		SecretKey:   record.GetString("secret_key"),
		Description: record.GetString("description"),
		Ext:         ext,

		Meta: meta,
	}

	return rs, nil
}
