package netmeta

import (
	"encoding/json"
)

type NodeMeta struct {
	Addr    string
	Quitted bool
}

type NetMeta map[string]NodeMeta

func NewNetMeta() NetMeta {
	return make(map[string]NodeMeta)
}

func NewLocalNetMeta(id, addr string) NetMeta {
	netMeta := NewNetMeta()
	netMeta[id] = Node{addr, false}
	return netMeta
}

func (netMeta NetMeta) Has(id string) bool {
	_, ok := netMeta[id]
	return ok
}

// return id, newNode, true if the update results in a change and "", nil, false otherwise
func (netMeta NetMeta) Update(id string, newNode NodeMeta) (string, NodeMeta, bool) {
	if node, ok := netMeta[id]; !ok || (!node.Quitted && newNode.Quitted) {
		netMeta[id] = newNode
		return id, newNode, true
	}
	return "", nil, false
}

// return the changes resulting from merge
func (netMeta NetMeta) Merge(netMeta2 NetMeta) (NetMeta, bool) {
	delta := NewNetMeta()
	for id, node := range netMeta2 {
		deltaId, deltaNode, changed := netMeta.Update(id, node)
		if changed {
			delta[deltaId] = deltaNode
		}
	}
	if len(delta) == 0 {
		return nil, false
	}
	return delta, true
}

func (netMeta NetMeta) ToJson() []byte {
	metaJson, _ := json.Marshal(netMeta)
	return metaJson
}

func (netMeta NetMeta) Copy() NetMeta {
	newMeta := NewNetMeta()
	for id, node := range netMeta {
		newMeta[id] = node
	}
	return newMeta
}
