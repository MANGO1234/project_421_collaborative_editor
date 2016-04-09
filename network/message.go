package network

import "encoding/json"

// message types of messages sent to existing connection
const (
	msgTypeVersionCheck  = "versioncheck" // no broadcast
	msgTypeSync          = "sync"         // no broadcast
	msgTypeNetMetaUpdate = "netmeta"      // broadcast
	msgTypeTreedocOp     = "treedocOp"    // broadcast
)

// TODO: for convenience, we are passing json around with possibly
// layers of mappings. This is inefficient. Should improve on this
// if we have time

// Message specifies the format of communication between nodes
// after connection establishes
type Message struct {
	Type    string
	Visited VisitedNodes
	Msg     []byte
}

func newBroadcastMessage(id, msgType string, content []byte) Message {
	return Message{
		msgType,
		newVisitedNodesWithSelf(id),
		content,
	}
}

func newNetMetaUpdateMsg(id string, delta NetMeta) Message {
	return newBroadcastMessage(id, msgTypeNetMetaUpdate, delta.toJson())
}

func newNetMetaUpdateMsgFromBytes(id string, delta []byte) Message {
	return newBroadcastMessage(id, msgTypeNetMetaUpdate, delta)
}

func newTreedocOpBroadcastMsg(id string, content []byte) Message {
	return newBroadcastMessage(id, msgTypeTreedocOp, content)
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

type VersionCheckMsgContent struct {
	Source        string
	NetworkMeta   NetMeta
	VersionVector []byte
}

func newVersionCheckMsgContentFromJson(contentJson []byte) (VersionCheckMsgContent, error) {
	var versionCheckMsgContent VersionCheckMsgContent
	err := json.Unmarshal(contentJson, &versionCheckMsgContent)
	return versionCheckMsgContent, err
}

func (content VersionCheckMsgContent) toJson() []byte {
	contentJson, _ := json.Marshal(content)
	return contentJson
}

//func newVersionCheckMsg(netMeta NetMeta, content []byte) Message {
//	msgContent := VersionCheckMsgContent{netMeta, content}
//	return newSyncOrCheckMessage(msgTypeVersionCheck, msgContent.toJson())
//}

func (msg *Message) toJson() []byte {
	msgJson, _ := json.Marshal(msg)
	return msgJson
}

func newMessageFromJson(msgJson []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(msgJson, &msg)
	return msg, err
}
