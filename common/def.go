package common

import (
	"encoding/binary"
	"github.com/satori/go.uuid"
)

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

func Uint32ToBytes(num uint32) []byte {
	newByte := make([]byte, 4)
	binary.BigEndian.PutUint32(newByte, num)
	return newByte
}

func BytesToUint32(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(bytes)
}

func UuidToBytes(id string) []byte {
	givenUUID, _ := uuid.FromString(id)
	bytesUUID := givenUUID.Bytes()
	return bytesUUID
}