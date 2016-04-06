package common

import "github.com/satori/go.uuid"

type SiteId [16]byte
type OperationId [20]byte

func (id SiteId) ToString() string {
	var idBytes []byte
	copy(idBytes, id[:])
	newUUID, _ := uuid.FromBytes(idBytes)
	return newUUID.String()
}

func StringToSiteId(id string) SiteId {
	var b [16]byte
	copy(b[:], id[:])
	return b
}
