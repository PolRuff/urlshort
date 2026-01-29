package audit

// Sink is an interface for audit log receivers (Observers)
type Sink interface {
	// Send sends an audit event to the receiver
	Send(event AuditEvent) error
}
