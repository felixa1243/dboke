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
	dbName string
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
	m.dbName = config.Database
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
		var stats bson.M
		err := db.RunCommand(ctx, bson.D{{Key: "collStats", Value: coll}}).Decode(&stats)
		
		rows := "Unknown"
		sizeStr := "Unknown"
		if err == nil {
			if count, ok := stats["count"]; ok {
				rows = fmt.Sprintf("%v", count)
			}
			if size, ok := stats["size"]; ok {
				// Type assertion for size could be int32 or int64 depending on BSON parsing
				var val int64
				switch v := size.(type) {
				case int32:
					val = int64(v)
				case int64:
					val = v
				case float64:
					val = int64(v)
				}
				
				if val < 1024 {
					sizeStr = fmt.Sprintf("%d B", val)
				} else if val < 1024*1024 {
					sizeStr = fmt.Sprintf("%d KB", val/1024)
				} else {
					sizeStr = fmt.Sprintf("%d MB", val/(1024*1024))
				}
			}
		}

		tables = append(tables, domain.TableMetadata{
			Name: coll,
			Type: "COLLECTION",
			Size: sizeStr, 
			Rows: rows,
		})
	}
	return tables, nil
}

func (m *MongoDBAdapter) GetTableSchema(ctx context.Context, database, table string) ([]domain.ColumnMetadata, error) {
	db := m.client.Database(database)
	coll := db.Collection(table)
	
	var doc bson.M
	err := coll.FindOne(ctx, bson.D{}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []domain.ColumnMetadata{}, nil
		}
		return nil, err
	}
	
	var columns []domain.ColumnMetadata
	for k, v := range doc {
		valType := "string"
		if v != nil {
			valType = fmt.Sprintf("%T", v)
		}
		
		columns = append(columns, domain.ColumnMetadata{
			Name: k,
			Type: valType,
			IsPrimaryKey: k == "_id",
			IsNullable: true,
		})
	}
	
	return columns, nil
}

func (m *MongoDBAdapter) FetchRows(ctx context.Context, query domain.SelectQuery) (*domain.ResultSet, error) {
	return nil, fmt.Errorf("FetchRows not fully implemented for MongoDB yet")
}

func (m *MongoDBAdapter) UpdateRow(ctx context.Context, update domain.UpdateQuery) (int64, error) {
	return 0, fmt.Errorf("UpdateRow not fully implemented for MongoDB yet")
}

func (m *MongoDBAdapter) ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error) {
	if m.dbName == "" {
		return nil, fmt.Errorf("no database selected for mongo execution")
	}
	
	db := m.client.Database(m.dbName)
	
	var cmd bson.D
	if err := bson.UnmarshalExtJSON([]byte(query), true, &cmd); err != nil {
		return nil, fmt.Errorf("query must be valid MongoDB Extended JSON (e.g. {\"find\": \"collection\"}): %w", err)
	}
	
	res := db.RunCommand(ctx, cmd)
	if res.Err() != nil {
		return nil, res.Err()
	}
	
	var result bson.M
	if err := res.Decode(&result); err != nil {
		return nil, err
	}
	
	var rows []map[string]interface{}
	
	if cursor, ok := result["cursor"].(bson.M); ok {
		if firstBatch, ok := cursor["firstBatch"].(bson.A); ok {
			for _, doc := range firstBatch {
				if mDoc, ok := doc.(bson.M); ok {
					rows = append(rows, mDoc)
				}
			}
		}
	} else {
		// Just return the command result as a single row
		rows = append(rows, result)
	}
	
	if len(rows) == 0 {
		return &domain.ResultSet{Columns: []string{}, Rows: []map[string]interface{}{}}, nil
	}
	
	colMap := make(map[string]bool)
	var columns []string
	for _, row := range rows {
		for k := range row {
			if !colMap[k] {
				colMap[k] = true
				columns = append(columns, k)
			}
		}
	}
	
	return &domain.ResultSet{Columns: columns, Rows: rows}, nil
}
