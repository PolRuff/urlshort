package audit

// AuditEvent represents a single audit log event
type AuditEvent struct {
	Timestamp int64  `json:"ts"`      // unix timestamp события
	Action    string `json:"action"`  // действие: shorten (создание) или follow (прохождение по ссылке)
	UserID    string `json:"user_id"` // идентификатор пользователя, если есть
	URL       string `json:"url"`     // оригинальный (не сокращенный) URL
}
