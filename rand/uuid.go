package uuid

import "github.com/google/uuid"

type UID string

func NewUUID() UID {
	return UID(uuid.NewString())
}
