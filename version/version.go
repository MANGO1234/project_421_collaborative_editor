package version

import (
	"github.com/satori/go.uuid"
)

type SiteId [16]byte
type VersionVector map[SiteId]uint32

const LESS_THAN = -1
const EQUAL = 0
const GREATER_THAN = 1
const CONFLICT = 2

func NewVector() VersionVector {
	return make(map[SiteId]uint32)
}

func (vector VersionVector) Get(id SiteId) uint32 {
	return vector[id]
}

func (vector VersionVector) Increment(id SiteId) {
	vector[id]++
}

func (vector VersionVector) IncrementTo(id SiteId, i uint32) {
	if vector[id] < i {
		vector[id] = i
	}
}

func (vector VersionVector) Merge(v2 VersionVector) {
	for k, v := range v2 {
		if vector[k] < v {
			vector[k] = v
		}
	}
}

func (vector VersionVector) Copy() VersionVector {
	newVector := NewVector()
	for k, v := range vector {
		newVector[k] = v
	}
	return newVector
}

func (v1 VersionVector) Compare(v2 VersionVector) int {
	r := EQUAL
	for k, v := range v1 {
		if v > v2[k] {
			if r == EQUAL {
				r = GREATER_THAN
			} else if r == LESS_THAN {
				return CONFLICT
			}
		} else if v < v2[k] {
			if r == EQUAL {
				r = LESS_THAN
			} else if r == GREATER_THAN {
				return CONFLICT
			}
		}
	}

	for k, v := range v2 {
		if v > v1[k] {
			if r == EQUAL {
				r = LESS_THAN
			} else if r == GREATER_THAN {
				return CONFLICT
			}
		} else if v < v1[k] {
			if r == EQUAL {
				r = GREATER_THAN
			} else if r == LESS_THAN {
				return CONFLICT
			}
		}
	}
	return r
}

func (id SiteId) ToString() string {
	newUUID, _ := uuid.FromBytes(id)
	return newUUID.String()
}

func NewSiteId(id string) SiteId {
	newUUID, _ := uuid.FromString(id)
	return newUUID.Bytes()
}

func (version *VersionVector) MarshalJSON([]byte, error) {
	newVector := make(map[string]uint32)
	for k, v := range version {
		newVector[k.ToString()] = v
	}
	return json.Marshal(newVector)
}

func (version *VersionVector) UnmarshalJSON(data []byte) error {
	newVector := make(map[string]uint32)
	if err := json.Unmarshal(data, &newVector); err != nil {
		return err
	}

	for k, v := range newVector {
		id := NewSiteId(k)
		version[id] = v
	}

	return nil
}
