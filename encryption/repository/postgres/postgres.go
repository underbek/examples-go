package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/underbek/examples-go/encryption/domain"
	gokitErrors "github.com/underbek/examples-go/errors"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
)

type postgres struct {
	conn goKitPgx.ExtContext
}

func New(conn goKitPgx.ExtContext) *postgres {
	return &postgres{conn: conn}
}

// GetEncryptorData returns which engine to use, you have to pass either encryptorType or id
func (p postgres) GetEncryptorData(ctx context.Context, encryptorType string, id string) (domain.EncryptorData, error) {

	builder := sq.Select(
		"id",
		"engine",
		"encryptor_type",
		"additional",
	).From("encryptors")

	if encryptorType != "" && id == "" {
		builder = builder.Where(sq.Eq{"encryptor_type": encryptorType})
	} else if id != "" && encryptorType == "" {
		builder = builder.Where(sq.Eq{"id": id})
	} else {
		return domain.EncryptorData{}, gokitErrors.Wrap(fmt.Errorf("encryptorType or id is required"), gokitErrors.TypeDatabase, "create query failed")
	}

	builder = builder.OrderBy("created_at DESC").Limit(1)

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return domain.EncryptorData{}, gokitErrors.Wrap(err, gokitErrors.TypeDatabase, "create sql string failed")
	}

	rows, err := p.conn.Query(ctx, query, args...)
	if err != nil {
		return domain.EncryptorData{}, gokitErrors.Wrap(err, gokitErrors.TypeDatabase, "query failed")
	}

	response, err := pgx.CollectOneRow[domain.EncryptorData](rows, pgx.RowToStructByName[domain.EncryptorData])
	if err != nil {
		return domain.EncryptorData{}, gokitErrors.Wrap(err, gokitErrors.TypeDatabase, "row scan failed")
	}

	return response, nil
}
