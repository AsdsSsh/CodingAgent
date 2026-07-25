package sandbox

import "strings"

// PermissionLevel is a monotonically increasing security level.
// Higher levels implicitly include all lower levels.
type PermissionLevel int

const (
	LevelNone     PermissionLevel = 0
	LevelRead     PermissionLevel = 1
	LevelWrite    PermissionLevel = 2
	LevelExecute  PermissionLevel = 3
	LevelDangerous PermissionLevel = 4
)

// Covers returns true if this level covers (is >=) the required level.
func (l PermissionLevel) Covers(required PermissionLevel) bool {
	return l >= required
}

// String returns the display name of the permission level.
func (l PermissionLevel) String() string {
	switch l {
	case LevelNone:
		return "NONE"
	case LevelRead:
		return "READ"
	case LevelWrite:
		return "WRITE"
	case LevelExecute:
		return "EXECUTE"
	case LevelDangerous:
		return "DANGEROUS"
	default:
		return "READ"
	}
}

// FromString parses a permission level from a config string.
func FromString(s string) PermissionLevel {
	switch strings.ToUpper(s) {
	case "NONE":
		return LevelNone
	case "READ":
		return LevelRead
	case "WRITE":
		return LevelWrite
	case "EXECUTE":
		return LevelExecute
	case "DANGEROUS":
		return LevelDangerous
	default:
		return LevelRead
	}
}
