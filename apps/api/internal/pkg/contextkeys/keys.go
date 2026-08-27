package contextkeys

type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	RoleKey      ContextKey = "role"
	SessionIDKey ContextKey = "session_id"
)
