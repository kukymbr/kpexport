package kinopoisk

import (
	"fmt"
	"strconv"
)

type UserID uint64

func (uid *UserID) String() string {
	return fmt.Sprintf("%d", *uid)
}

func (uid *UserID) ToURL() string {
	return Host + "/user/" + uid.String()
}

// Set sets a value.
// Used for the CLI flag parsing.
func (uid *UserID) Set(val string) error {
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert '%s' into user ID: %w", val, err)
	}

	*uid = UserID(id)

	return nil
}

// Type is a name of the type, used for the CLI flag info.
func (uid *UserID) Type() string {
	return "UserID"
}
