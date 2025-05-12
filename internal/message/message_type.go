package message

type MessageType string

const (
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
	MessageTypeStart   MessageType = "start"
	MessageTypeStarted MessageType = "started"

	MessageTypeRoll   MessageType = "roll"
	MessageTypeRolled MessageType = "rolled"
	MessageTypeHold   MessageType = "hold"
	MessageTypeUndo   MessageType = "undo"
	MessageTypeLock   MessageType = "lock"
	MessageTypeBank   MessageType = "bank"

	MessageTypeMe MessageType = "me"
)
