package network

import (
	"encoding/json"
	"sync"
)

type NodeMeta struct {
	Addr string
	Left bool
}

type NetMeta map[string]NodeMeta

type nodePool struct {
	netMetaMutex     sync.RWMutex
	poolMutex        sync.Mutex
	netMeta          NetMeta
	connectedPool    map[string]*node
	disconnectedPool map[string]*node
}

func newNodePool() *nodePool {
	var np nodePool
	np.netMeta = newNetMeta()
	np.connectedPool = make(map[string]*node)
	np.disconnectedPool = make(map[string]*node)
	return &np
}

func newNetMeta() NetMeta {
	return make(map[string]NodeMeta)
}

func newQuitNetMeta(id, addr string) NetMeta {
	netMeta := newNetMeta()
	netMeta[id] = NodeMeta{addr, true}
	return netMeta
}

func newJoinNetMeta(id, addr string) NetMeta {
	netMeta := newNetMeta()
	netMeta[id] = NodeMeta{addr, false}
	return netMeta
}

func (netMeta NetMeta) ToJson() []byte {
	netMetaJson, _ := json.Marshal(netMeta)
	return netMetaJson
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

func (np *nodePool) getLatestNetmetaCopy() NetMeta {
	newMeta := newNetMeta()
	np.netMetaMutex.RLock()
	for id, node := range np.netMeta {
		newMeta[id] = node
	}
	np.netMetaMutex.RUnlock()
	return newMeta
}
