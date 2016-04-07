package network

import "encoding/json"

// message types of messages sent to existing connection
const (
	msgTypeVersionCheck  = "versioncheck" // no broadcast
	msgTypeSync          = "sync"         // no broadcast
	msgTypeNetMetaUpdate = "netmeta"      // broadcast
	msgTypeTreedocOp     = "treedocOp"    // broadcast
)

// Message specifies the format of communication between nodes
// after connection establishes
type Message struct {
	Type    string
	Visited map[string]struct{}
	Msg     []byte
}

func newBroadcastMessage(msgType string, content []byte) Message {
	return Message{
		msgType,
		make(map[string]struct{}),
		content,
	}
}

func newNetMetaUpdateMsg(delta NetMeta) Message {
	return newBroadcastMessage(msgTypeNetMetaUpdate, delta.ToJson())
}

func newNetMetaUpdateMsgFromBytes(delta []byte) Message {
	return newBroadcastMessage(msgTypeNetMetaUpdate, delta)
}

func newTreedocOpBroadcastMsg(content []byte) Message {
	return newBroadcastMessage(msgTypeTreedocOp, content)
}

func newSyncOrCheckMessage(msgType string, content []byte) Message {
	return Message{
		msgType,
		nil,
		content,
	}
}

func newSyncMessage(content []byte) Message {
	return newSyncOrCheckMessage(msgTypeSync, content)
}

func newVersionCheckMsg(netMeta NetMeta, content []byte) Message {
	contentJson := VersionCheckMsgContent{netMeta, content}.toJson()
	return newSyncOrCheckMessage(msgTypeVersionCheck, contentJson)

}

type VersionCheckMsgContent struct {
	NetworkMeta   NetMeta
	VersionVector []byte
}

func (content VersionCheckMsgContent) toJson() []byte {
	contentJson, _ := json.Marshal(content)
	return contentJson
}
