package database

import (
	"context"
	"fmt"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBAdapter struct {
	client *mongo.Client
}

func NewMongoDBAdapter() ports.IDBAdapter {
	return &MongoDBAdapter{}
}

func init() {
	RegisterAdapter("mongodb", NewMongoDBAdapter)
}

func (m *MongoDBAdapter) Connect(ctx context.Context, config domain.DBConnectionConfig) error {
	// Construct connection string
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=admin", 
		config.User, config.Password, config.Host, config.Port)
	
	// Handle unauthenticated connections
	if config.User == "" {
		uri = fmt.Sprintf("mongodb://%s:%d", config.Host, config.Port)
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create mongo client: %w", err)
	}

	m.client = client
	return m.Ping(ctx)
}

func (m *MongoDBAdapter) Disconnect() error {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return m.client.Disconnect(ctx)
	}
	return nil
}

func (m *MongoDBAdapter) Ping(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("mongo client not initialized")
	}
	return m.client.Ping(ctx, readpref.Primary())
}

func (m *MongoDBAdapter) GetDatabases(ctx context.Context) ([]string, error) {
	names, err := m.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	return names, nil
}

func (m *MongoDBAdapter) GetTables(ctx context.Context, database string) ([]domain.TableMetadata, error) {
	db := m.client.Database(database)
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var tables []domain.TableMetadata
	for _, coll := range collections {
		tables = append(tables, domain.TableMetadata{
			Name: coll,
			Type: "COLLECTION",
			Size: "Unknown", 
			Rows: "Unknown",
		})
	}
	return tables, nil
}

func (m *MongoDBAdapter) GetTableSchema(ctx context.Context, database, table string) ([]domain.ColumnMetadata, error) {
	return nil, fmt.Errorf("schema inference not fully implemented for MongoDB yet")
}

func (m *MongoDBAdapter) FetchRows(ctx context.Context, query domain.SelectQuery) (*domain.ResultSet, error) {
	return nil, fmt.Errorf("FetchRows not fully implemented for MongoDB yet")
}

func (m *MongoDBAdapter) UpdateRow(ctx context.Context, update domain.UpdateQuery) (int64, error) {
	return 0, fmt.Errorf("UpdateRow not fully implemented for MongoDB yet")
}

func (m *MongoDBAdapter) ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error) {
	return nil, fmt.Errorf("ExecuteRaw not fully implemented for MongoDB yet")
}
