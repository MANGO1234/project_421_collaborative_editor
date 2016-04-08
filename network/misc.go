package network

type VisitedNodes map[string]struct{}

func newVisitedNodesFromNodeList(nodeList []*node) VisitedNodes {
	v := newVisitedNodes()
	for _, n := range nodeList {
		v[n.id] = struct{}{}
	}
	return v
}

func newVisitedNodes() VisitedNodes {
	return make(map[string]struct{})
}

func newVisitedNodesWithSelf(id string) VisitedNodes {
	v := newVisitedNodes()
	v[id] = struct{}{}
	return v
}

func (v VisitedNodes) copyVisitedNodes() VisitedNodes {
	vCopy := newVisitedNodes()
	for id, _ := range v {
		vCopy[id] = struct{}{}
	}
	return vCopy
}

func (v VisitedNodes) addAll(v2 VisitedNodes) {
	for id, _ := range v2 {
		v[id] = struct{}{}
	}
}

func (v VisitedNodes) addAllFromNodeList(nodeList []*node) {
	for _, n := range nodeList {
		v[n.id] = struct{}{}
	}
}

func (v VisitedNodes) has(id string) bool {
	_, ok := v[id]
	return ok
}
