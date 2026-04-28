package domain

// AccountStatus mirrors auth_accounts.status.
type AccountStatus string

const (
	StatusNew                 AccountStatus = "new"
	StatusPendingVerification AccountStatus = "pending_verification"
	StatusActive              AccountStatus = "active"
	StatusBlocked             AccountStatus = "blocked"
	StatusDeleted             AccountStatus = "deleted"
)

// Role mirrors auth_accounts.role.
type Role string

const (
	RoleUser   Role = "user"
	RoleAdmin  Role = "admin"
	RoleSystem Role = "system"
)
