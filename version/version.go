package version

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
	if len(v2) > len(v1) {
		r := v2.Compare(v1)
		if r == LESS_THAN {
			return GREATER_THAN
		} else if r == GREATER_THAN {
			return LESS_THAN
		}
		return r
	}

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
	return r
}
