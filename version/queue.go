package version

type QueueElem struct {
	Vector    VersionVector
	Id        SiteId
	Version   uint32
	Operation interface{}
}

type VectorQueue struct {
	queue  []QueueElem
	vector VersionVector
}

func NewQueue() *VectorQueue {
	return &VectorQueue{make([]QueueElem, 0, 4), NewVector()}
}

func (queue *VectorQueue) Size() int {
	return len(queue.queue)
}

func (queue *VectorQueue) Vector() VersionVector {
	return queue.vector.Copy()
}

// enqueue an operation and returns list of operation that's ready
func (queue *VectorQueue) Enqueue(elem QueueElem) []QueueElem {
	if queue.vector.Get(elem.Id) == elem.Version-1 {
		compare := queue.vector.Compare(elem.Vector)
		if compare == GREATER_THAN || compare == EQUAL {
			result := make([]QueueElem, 1, 4)
			queue.vector.Increment(elem.Id)
			result[0] = elem
			result, offset := dequeHelper(result, queue, len(queue.queue))
			queue.queue = queue.queue[:len(queue.queue)-offset]
			return result
		}
	}
	queue.queue = append(queue.queue, elem)
	return nil
}

func dequeHelper(result []QueueElem, queue *VectorQueue, upto int) ([]QueueElem, int) {
	q := queue.queue
	v := queue.vector
	offset := 0
	for i := 0; i < upto; i++ {
		if offset != 0 {
			q[i-offset] = q[i]
		}
		// remove operation that already exists
		if v.Get(q[i].Id) >= q[i].Version {
			offset += 1
			continue
		}
		compare := v.Compare(q[i].Vector)
		if compare == GREATER_THAN || compare == EQUAL {
			v.Increment(q[i].Id)
			result = append(result, q[i])
			var offsetp int
			result, offsetp = dequeHelper(result, queue, i-offset)
			offset += offsetp + 1
		}
	}
	return result, offset
}
