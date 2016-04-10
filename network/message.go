package network

import "encoding/json"

const (
	MSG_TYPE_VERSION_CHECK   = "versioncheck" // no broadcast
	MSG_TYPE_NET_META_UPDATE = "netmeta"      // broadcast
	MSG_TYPE_TREEDOC_OP      = "treedocOp"    // broadcast
)

// Message specifies the format of communication between nodes
// after connection establishes
type Message struct {
	Type    string
	Visited VisitedNodes
	Msg     []byte
}

func NewBroadcastMessage(id, msgType string, content []byte) Message {
	return Message{
		msgType,
		newVisitedNodesWithSelf(id),
		content,
	}
}

func newNetMetaUpdateMsg(id string, delta NetMeta) Message {
	return NewBroadcastMessage(id, MSG_TYPE_NET_META_UPDATE, delta.toJson())
}

func newSyncOrCheckMessage(msgType string, content []byte) Message {
	return Message{
		msgType,
		nil,
		content,
	}
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

func (msg *Message) toJson() []byte {
	msgJson, _ := json.Marshal(msg)
	return msgJson
}

func newMessageFromJson(msgJson []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(msgJson, &msg)
	return msg, err
}
