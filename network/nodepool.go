package network

import (
	"sync"
)

type nodePool struct {
	netMetaMutex sync.RWMutex
	netMeta      NetMeta
	poolMutex    sync.RWMutex
	pool         map[string]*node
}

func newNodePool() *nodePool {
	var np nodePool
	np.netMeta = newNetMeta()
	np.pool = make(map[string]*node)
	return &np
}

func (np *nodePool) has(id string) bool {
	_, ok := np.netMeta[id]
	return ok
}

func (np *nodePool) handleNewSession(s *session) {
	np.netMetaMutex.Lock()
	np.netMeta[s.id] = NodeMeta{s.manager.addr, false}
	np.netMetaMutex.Unlock()
}

// func (np *nodepool) handleEndSession(s *session) {

// }

func (np *nodePool) getConnectedNodes() []*node {
	np.poolMutex.RLock()
	defer np.poolMutex.RUnlock()
	nodes := make([]*node, 0)
	for _, v := range np.pool {
		if v.state == nodeStateConnected {
			nodes = append(nodes, v)
		}
	}
	return nodes
}

func (np *nodePool) getAllNodes() []*node {
	np.poolMutex.RLock()
	defer np.poolMutex.RUnlock()
	return getNodeListFromMap(np.pool)
}

func getNodeListFromMap(nodeMap map[string]*node) []*node {
	length := len(nodeMap)
	nodes := make([]*node, length, length)
	i := 0
	for _, n := range nodeMap {
		nodes[i] = n
		i++
	}
	return nodes
}

func (np *nodePool) disconnectAllNodes() {
	np.poolMutex.RLock()
	for _, n := range np.pool {
		// TODO: check locking issues
		n.close()
	}
	np.poolMutex.RUnlock()
}

func (np *nodePool) removeNodeFromPool(id string) {
	np.poolMutex.Lock()
	if n, ok := np.pool[id]; ok {
		n.leave()
		delete(np.pool, id)
	}
	np.poolMutex.Unlock()
}

func (np *nodePool) addNodeToPool(id string, nodeMeta NodeMeta) *node {
	np.poolMutex.Lock()
	n := newNodeFromIdNodeMeta(id, nodeMeta)
	np.pool[id] = n
	np.poolMutex.Unlock()
	return n
}

// return any new nodes to be handled and updates applied and whether any changes occurred
func (np *nodePool) applyReceivedUpdates(updates NetMeta) (nodeList []*node, delta NetMeta, changed bool) {
	nodeList = make([]*node, 0)
	np.netMetaMutex.Lock()
	delta, changed = np.netMeta.merge(updates)
	np.netMetaMutex.Unlock()
	for id, n := range delta {
		if n.Left {
			np.removeNodeFromPool(id)
		} else {
			nodeList = append(nodeList, np.addNodeToPool(id, n))
		}
	}
	return
}

func (np *nodePool) forceNodeQuit(n *node) {

}

//func (np *nodePool) sendMs

func (np *nodePool) getLatestNetMetaJson() []byte {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta.toJson()
}

func (np *nodePool) getLatestNetMetaJsonPrettyPrint() []byte {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta.toJsonPrettyPrint()
}

func (np *nodePool) getLatestNetMetaCopy() NetMeta {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta.copy()
}
