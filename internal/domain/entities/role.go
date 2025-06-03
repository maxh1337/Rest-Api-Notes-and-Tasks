package entities

import (
	"database/sql/driver"
	"fmt"
)

type RoleType string

const (
	RoleAdmin     RoleType = "admin"
	RoleUser      RoleType = "user"
	RoleModerator RoleType = "moderator"
	RoleGuest     RoleType = "guest"
)

func (r RoleType) String() string {
	return string(r)
}

func (r RoleType) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser, RoleModerator, RoleGuest:
		return true
	default:
		return false
	}
}

func (r RoleType) Value() (driver.Value, error) {
	if !r.IsValid() {
		return nil, fmt.Errorf("invalid role type: %s", r)
	}
	return string(r), nil
}

func (r *RoleType) Scan(value interface{}) error {
	if value == nil {
		*r = RoleGuest
		return nil
	}

	switch s := value.(type) {
	case string:
		*r = RoleType(s)
	case []byte:
		*r = RoleType(s)
	default:
		return fmt.Errorf("cannot scan %T into RoleType", value)
	}

	if !r.IsValid() {
		return fmt.Errorf("invalid role type: %s", *r)
	}

	return nil
}
