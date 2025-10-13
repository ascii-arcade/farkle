package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	URI      string
	Database string

	client *mongo.Client
	ctx    context.Context
}

var db *MongoDB

func Setup(ctx context.Context, uri, database string) error {
	db = &MongoDB{
		URI:      uri,
		Database: database,
		ctx:      ctx,
	}
	return db.Connect()
}

func GetDB() *MongoDB {
	return db
}

func (m *MongoDB) Connect() (err error) {
	m.client, err = mongo.Connect(m.ctx, options.Client().ApplyURI(m.URI))
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) Close() error {
	return m.client.Disconnect(m.ctx)
}
