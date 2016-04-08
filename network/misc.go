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

func copyVisitedNodes(v VisitedNodes) VisitedNodes {
	vCopy := newVisitedNodes()
	for id, _ := range v {
		vCopy[id] = struct{}{}
	}
	return vCopy
}

func (v VisitedNodes) merge(v2 VisitedNodes ) {
	for id, _:=range v2 {
		v[id] = struct{}{}
	}
}