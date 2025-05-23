package message

type MessageType string

const (
	MessageTypeList  MessageType = "list"
	MessageTypeError MessageType = "error"

	MessageTypeUpdate  MessageType = "update"
	MessageTypeUpdated MessageType = "updated"

	MessageTypeLeave MessageType = "leave"
	MessageTypeStart MessageType = "start"

	MessageTypeRoll   MessageType = "roll"
	MessageTypeRolled MessageType = "rolled"
	MessageTypeHold   MessageType = "hold"
	MessageTypeUndo   MessageType = "undo"
	MessageTypeLock   MessageType = "lock"
	MessageTypeBank   MessageType = "bank"

	MessageTypeMe MessageType = "me"
)
