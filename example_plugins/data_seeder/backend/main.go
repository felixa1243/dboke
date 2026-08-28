package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/hashicorp/go-plugin"

	"dboke-plugins/shared"
)

// DataSeeder implements the shared.Feature interface
type DataSeeder struct{}

// payloadSchema represents the JSON payload expected from the UI
type payloadSchema struct {
	Table   string            `json:"table"`
	Columns map[string]string `json:"columns"`
	Count   int               `json:"count"`
}

func (s *DataSeeder) BuildQuery(visualPayload string) (string, error) {
	var req payloadSchema
	if err := json.Unmarshal([]byte(visualPayload), &req); err != nil {
		return "", fmt.Errorf("failed to parse payload: %v", err)
	}

	if req.Table == "" {
		return "", fmt.Errorf("table name is required")
	}
	if len(req.Columns) == 0 {
		return "", fmt.Errorf("no columns mapped")
	}
	if req.Count <= 0 {
		req.Count = 10 // default
	}

	gofakeit.Seed(0) // Random seed

	var colNames []string
	for col := range req.Columns {
		colNames = append(colNames, col)
	}

	var valuesStrings []string

	for i := 0; i < req.Count; i++ {
		var rowValues []string
		for _, colName := range colNames {
			colType := req.Columns[colName]
			val := generateFakeData(colType)
			rowValues = append(rowValues, val)
		}
		valuesStrings = append(valuesStrings, "("+strings.Join(rowValues, ", ")+")")
	}

	colsStr := strings.Join(colNames, ", ")
	valsStr := strings.Join(valuesStrings, ", ")
	
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;", req.Table, colsStr, valsStr)
	return query, nil
}

func generateFakeData(fakeType string) string {
	var raw interface{}
	switch fakeType {
	case "uuid":
		raw = gofakeit.UUID()
	case "number":
		raw = gofakeit.Number(1, 1000)
	case "money":
		raw = gofakeit.Price(1.0, 1000.0)
	case "person":
		raw = gofakeit.Name()
	case "email":
		raw = gofakeit.Email()
	case "date":
		raw = gofakeit.Date().Format("2006-01-02 15:04:05")
	case "word":
		raw = gofakeit.Word()
	case "address":
		raw = gofakeit.Address().Address
	case "company":
		raw = gofakeit.Company()
	default:
		raw = gofakeit.Word() // fallback
	}

	// Format for SQL (strings get quotes, numbers don't)
	switch v := raw.(type) {
	case int, float64, int64:
		return fmt.Sprintf("%v", v)
	default:
		// Escape single quotes in strings
		escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

func (s *DataSeeder) GetFrontendComponent() (string, error) {
	// Our Next.js hack means we don't need to return raw JS here; Next.js serves the frontend natively.
	return "", nil
}

func main() {
	seeder := &DataSeeder{}
	
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"feature": &shared.FeaturePlugin{Impl: seeder},
		},
		// Note: The plugin manager allows ProtocolNetRPC
		// So we do not force GRPC here.
	})
}
