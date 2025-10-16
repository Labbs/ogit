package domain

type MemberType string
type AccessType string

type Member struct {
	Id     string
	Type   MemberType
	Access AccessType
}

// AccessTypeViewer is the access type viewer
const (
	AccessTypeNone       AccessType = "none"
	AccessTypeReporter   AccessType = "reporter"
	AccessTypeDeveloper  AccessType = "developer"
	AccessTypeMaintainer AccessType = "maintainer"
	AccessTypeOwner      AccessType = "owner"
)

// MemberType is the type of member
const (
	MemberTypeUser  MemberType = "user"
	MemberTypeGroup MemberType = "group"
)

// Level returns the level of the AccessType
func (a AccessType) Level() int {
	switch a {
	case AccessTypeNone:
		return 0
	case AccessTypeReporter:
		return 1
	case AccessTypeDeveloper:
		return 2
	case AccessTypeMaintainer:
		return 3
	case AccessTypeOwner:
		return 4
	default:
		return 0
	}
}

// MaxAccessType returns the maximum AccessType between two AccessTypes
func MaxAccessType(a1, a2 AccessType) AccessType {
	if a1.Level() > a2.Level() {
		return a1
	}
	return a2
}

// Verify if the AccessType has read, write or admin permissions
func (a AccessType) CanRead() bool {
	return a.Level() >= AccessTypeReporter.Level()
}

func (a AccessType) CanWrite() bool {
	return a.Level() >= AccessTypeDeveloper.Level()
}

func (a AccessType) CanAdmin() bool {
	return a.Level() >= AccessTypeMaintainer.Level()
}
