package network

import (
	"encoding/json"
	"sync"
)

type nodePool struct {
	netMetaMutex          sync.RWMutex
	netMeta               NetMeta
	connectedPoolMutex    sync.RWMutex
	connectedPool         map[string]*node
	disconnectedPoolMutex sync.RWMutex
	disconnectedPool      map[string]*node
}

func newNodePool() *nodePool {
	var np nodePool
	np.netMeta = newNetMeta()
	np.connectedPool = make(map[string]*node)
	np.disconnectedPool = make(map[string]*node)
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

func (np *nodePool) getDisconnectedNodes() []*node {
	np.disconnectedPoolMutex.RLock()
	defer np.disconnectedPoolMutex.RUnlock()
	return getNodeListFromMap(np.disconnectedPool)
}

func (np *nodePool) getConnectedNodes() []*node {
	np.connectedPoolMutex.RLock()
	defer np.connectedPoolMutex.RUnlock()
	return getNodeListFromMap(np.connectedPool)
}

func getNodeListFromMap(nodeMap map[string]*node) []*node {
	length := len(nodeMap)
	nodes := make([]*node, length, length)
	i := 0
	for _, v := range nodeMap {
		nodes[i] = v
		i++
	}
	return nodes
}

// return true if the update results in a change and false otherwise
func (np *nodePool) applyReceivedUpdate(id string, newNode NodeMeta) bool {
	// TODO
	if node, ok := np.netMeta[id]; !ok || (!node.Left && newNode.Left) {
		np.netMeta[id] = newNode
		return true
	}
	return false
}

// return the updates applied and whether any changes occurred
func (np *nodePool) applyReceivedUpdates(updates NetMeta) (NetMeta, bool) {
	delta := newNetMeta()
	for id, node := range updates {
		changed := np.applyReceivedUpdate(id, node)
		if changed {
			delta[id] = node
		}
	}
	if len(delta) == 0 {
		return nil, false
	}
	return delta, true
}

func (np *nodePool) forceNodeQuit(n *node) {

}

//func (np *nodePool) sendMs

func (np *nodePool) getLatestNetMetaJson() []byte {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta.ToJson()
}

func (np *nodePool) getLatestNetMetaJsonPrettyPrint() []byte {
	np.netMetaMutex.RLock()
	metaJson, _ := json.MarshalIndent(np.netMeta, "", "    ")
	np.netMetaMutex.RUnlock()
	return metaJson
}

func (np *nodePool) getLatestNetMetaCopy() NetMeta {
	newMeta := newNetMeta()
	np.netMetaMutex.RLock()
	for id, node := range np.netMeta {
		newMeta[id] = node
	}
	np.netMetaMutex.RUnlock()
	return newMeta
}
