package notification

type NotificationPriority string

const (
	LowPriority    NotificationPriority = "low"
	NormalPriority NotificationPriority = "normal"
	HighPriority   NotificationPriority = "high"
)

type NotificationPayload struct {
	Title    string
	Body     string
	Priority NotificationPriority
}
