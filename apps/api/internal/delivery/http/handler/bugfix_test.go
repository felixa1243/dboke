package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/core/services"
	"dboke-api/internal/infrastructure/security"

	"github.com/go-chi/chi/v5"
)

func TestIsSelectQuery(t *testing.T) {
	cases := map[string]bool{
		"SELECT 1":                      true,
		"select * from t":               true,
		"sElEcT * from t":               true,
		"  \n\tSELECT 1":                true,
		"-- comment\nSELECT 1":          true,
		"/* c */ select 1":              true,
		"SELECT\n*":                     true,
		"SELECT(1)":                     true,
		"WITH x AS (SELECT 1) SELECT 1": false,
		"INSERT INTO t VALUES (1)":      false,
		"selection":                     false,
		"-- only a comment":             false,
		"/* unterminated SELECT":        false,
		"UPDATE t SET a = 1 -- select":  false,
	}
	for q, want := range cases {
		if got := isSelectQuery(q); got != want {
			t.Errorf("isSelectQuery(%q) = %v, want %v", q, got, want)
		}
	}
}

type recordingAdapter struct {
	mockDBAdapter
	lastQuery string
}

func (r *recordingAdapter) ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error) {
	r.lastQuery = query
	return &domain.ResultSet{Columns: []string{"a"}, Rows: []map[string]interface{}{{"a": 1}}}, nil
}

func TestDBHandler_ExecuteQuery_PaginatedFlag(t *testing.T) {
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)
	store := newMockSessionStore()
	cfg, _ := json.Marshal(domain.DBConnectionConfig{Host: "h", Port: 1})
	enc, _ := security.Encrypt(string(cfg), masterKey)
	store.CreateSession(context.Background(), &domain.Session{
		ID: "s1", DBType: "pgsql", EncryptedDBConfig: enc, ExpiresAt: time.Now().Add(time.Hour),
	})
	adapter := &recordingAdapter{}
	dbSvc := services.NewDBService(store, map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter { return adapter },
	})
	h := NewDBHandler(dbSvc)

	for _, tc := range []struct {
		query     string
		paginated bool
	}{
		{"  select * from users", true},
		{"INSERT INTO t VALUES (1) RETURNING *", false},
	} {
		body, _ := json.Marshal(map[string]interface{}{"query": tc.query, "limit": 10})
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)).WithContext(contextWithSession("s1"))
		w := httptest.NewRecorder()
		h.ExecuteQuery(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%q: status %d: %s", tc.query, w.Code, w.Body.String())
		}
		var resp struct {
			Paginated bool                     `json:"paginated"`
			Rows      []map[string]interface{} `json:"rows"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Paginated != tc.paginated {
			t.Errorf("%q: paginated = %v, want %v", tc.query, resp.Paginated, tc.paginated)
		}
		if len(resp.Rows) != 1 {
			t.Errorf("%q: result set fields missing from response: %s", tc.query, w.Body.String())
		}
		wrapped := len(adapter.lastQuery) > 0 && adapter.lastQuery[:len("SELECT * FROM (")] == "SELECT * FROM ("
		if wrapped != tc.paginated {
			t.Errorf("%q: wrapped = %v, want %v (sent %q)", tc.query, wrapped, tc.paginated, adapter.lastQuery)
		}
	}
}

func withPluginID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestPluginHandler_RejectsTraversalIDs(t *testing.T) {
	// Run from a temp working dir so a regression cannot delete real files
	tmp := t.TempDir()
	cwd := filepath.Join(tmp, "apps", "api")
	os.MkdirAll(cwd, 0755)
	canary := filepath.Join(tmp, "apps", "canary")
	os.WriteFile(canary, []byte("x"), 0644)
	orig, _ := os.Getwd()
	os.Chdir(cwd)
	defer os.Chdir(orig)

	h := NewPluginHandler(services.NewPluginManager(), nil)
	for _, id := range []string{"..", ".", "../..", "a/../.."} {
		w := httptest.NewRecorder()
		h.DeletePlugin(w, withPluginID(httptest.NewRequest(http.MethodDelete, "/", nil), id))
		if w.Code != http.StatusBadRequest {
			t.Errorf("DeletePlugin(%q) status = %d, want 400", id, w.Code)
		}
		w = httptest.NewRecorder()
		h.TogglePlugin(w, withPluginID(httptest.NewRequest(http.MethodPost, "/", nil), id))
		if w.Code != http.StatusBadRequest {
			t.Errorf("TogglePlugin(%q) status = %d, want 400", id, w.Code)
		}
	}
	if _, err := os.Stat(canary); err != nil {
		t.Fatalf("file outside plugins dir was removed: %v", err)
	}
}

func zipUploadRequest(t *testing.T, metaJSON string) *http.Request {
	t.Helper()
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, _ := zw.Create("meta.json")
	f.Write([]byte(metaJSON))
	zw.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("executable", "plugin.zip")
	part.Write(zipBuf.Bytes())
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/plugins/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestPluginHandler_Upload_DisabledByDefault(t *testing.T) {
	t.Setenv("DBOKE_ENABLE_PLUGIN_UPLOAD", "")
	h := NewPluginHandler(services.NewPluginManager(), nil)
	w := httptest.NewRecorder()
	h.UploadPlugin(w, zipUploadRequest(t, `{"id":"ok"}`))
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestPluginHandler_Upload_RejectsTraversalMetaID(t *testing.T) {
	t.Setenv("DBOKE_ENABLE_PLUGIN_UPLOAD", "true")
	tmp := t.TempDir()
	cwd := filepath.Join(tmp, "apps", "api")
	os.MkdirAll(cwd, 0755)
	orig, _ := os.Getwd()
	os.Chdir(cwd)
	defer os.Chdir(orig)

	h := NewPluginHandler(services.NewPluginManager(), nil)
	w := httptest.NewRecorder()
	h.UploadPlugin(w, zipUploadRequest(t, `{"id":"../../escape"}`))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	entries, _ := os.ReadDir(filepath.Join(tmp, "apps", "plugins"))
	if len(entries) != 0 {
		t.Errorf("temp upload not cleaned up: %v", entries)
	}
}

func TestAuthHandler_Login_DisallowedHost(t *testing.T) {
	t.Setenv("DBOKE_ALLOWED_DB_HOSTS", "db.internal, localhost")
	connected := false
	authSvc := services.NewAuthService(newMockSessionStore(), map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter { connected = true; return &mockDBAdapter{} },
	})
	h := NewAuthHandler(authSvc, nil)
	body, _ := json.Marshal(LoginRequest{DbType: "pgsql", Host: "attacker.example", Port: "5432"})
	w := httptest.NewRecorder()
	h.Login(w, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)))
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
	if connected {
		t.Error("adapter was created for a disallowed host")
	}
	if !isAllowedDBHost("DB.INTERNAL") {
		t.Error("allowlist should be case-insensitive and trim spaces")
	}
}

func TestAuthHandler_Logout_ClosesConnections(t *testing.T) {
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)
	store := newMockSessionStore()
	cfg, _ := json.Marshal(domain.DBConnectionConfig{Host: "h", Port: 1})
	enc, _ := security.Encrypt(string(cfg), masterKey)
	store.CreateSession(context.Background(), &domain.Session{
		ID: "abcd", DBType: "pgsql", EncryptedDBConfig: enc, ExpiresAt: time.Now().Add(time.Hour),
	})
	adapter := &mockDBAdapter{}
	factories := map[string]func() ports.IDBAdapter{"pgsql": func() ports.IDBAdapter { return adapter }}
	dbSvc := services.NewDBService(store, factories)
	if _, err := dbSvc.GetAdapter(context.Background(), "abcd", ""); err != nil {
		t.Fatal(err)
	}

	h := NewAuthHandler(services.NewAuthService(store, factories), dbSvc)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "abcd"})
	h.Logout(httptest.NewRecorder(), req)

	if adapter.Connected {
		t.Error("database connection still open after logout")
	}
}
