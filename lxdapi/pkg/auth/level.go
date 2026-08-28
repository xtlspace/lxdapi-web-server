package auth

type APILevel int

const (
	LevelSystem    APILevel = iota
	LevelAdmin
	LevelContainer
)

func (l APILevel) String() string {
	switch l {
	case LevelSystem:
		return "system"
	case LevelAdmin:
		return "admin"
	case LevelContainer:
		return "container"
	default:
		return "unknown"
	}
}

func (l APILevel) CanAccess(required APILevel) bool {
	return l <= required
}

