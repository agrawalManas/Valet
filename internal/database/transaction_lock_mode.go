package database

type LockMode int

const (
	NoLock LockMode = iota
	ForShare
	ForUpdate
)

func (m LockMode) Clause() string {
	switch m {
	case ForShare:
		return " FOR SHARE"
	case ForUpdate:
		return " FOR UPDATE"
	}
	return ""
}
