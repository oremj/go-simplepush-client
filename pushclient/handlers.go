package pushclient

type PushHandler interface {
	RegisterHandler(*RegisterResponse)
	NotificationHandler(*Notification)
}
