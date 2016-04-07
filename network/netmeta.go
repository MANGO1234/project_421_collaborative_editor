package network

import "encoding/json"

type NodeMeta struct {
	Addr string
	Left bool
}

type NetMeta map[string]NodeMeta

func newNetMeta() NetMeta {
	return make(map[string]NodeMeta)
}

func newQuitNetMeta(id, addr string) NetMeta {
	netMeta := newNetMeta()
	netMeta[id] = NodeMeta{addr, true}
	return netMeta
}

func newNetMetaFromJson(netMetaJson []byte) (NetMeta, error) {
	var netMeta NetMeta
	err := json.Unmarshal(netMetaJson, &netMeta)
	return netMeta, err
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
