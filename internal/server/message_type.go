package server

type MessageType string

const (
	MessageTypePing  MessageType = "ping"
	MessageTypeList  MessageType = "list"
	MessageTypeError MessageType = "error"

	MessageTypeCreate  MessageType = "create"
	MessageTypeDelete  MessageType = "delete"
	MessageTypeUpdate  MessageType = "update"
	MessageTypeCreated MessageType = "created"
	MessageTypeDeleted MessageType = "deleted"
	MessageTypeUpdated MessageType = "updated"
	MessageTypeJoin    MessageType = "join"
	MessageTypeLeave   MessageType = "leave"

	MessageTypeMe MessageType = "me"
)
