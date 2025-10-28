package shared

// EmailJob represents an email sending job
type EmailJob struct {
	ID        int    `json:"id"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Priority  string `json:"priority"` // high, normal, low
}

// QueueName implements the goqueue.QueueMessage interface
func (e *EmailJob) QueueName() string {
	return "email-jobs"
}

// ImageProcessingJob represents an image processing job
type ImageProcessingJob struct {
	ID        int    `json:"id"`
	ImageURL  string `json:"image_url"`
	Operation string `json:"operation"` // resize, crop, filter
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// QueueName implements the goqueue.QueueMessage interface
func (i *ImageProcessingJob) QueueName() string {
	return "image-jobs"
}

// NotificationJob represents a push notification job
type NotificationJob struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
	Type    string `json:"type"` // push, sms, slack
}

// QueueName implements the goqueue.QueueMessage interface
func (n *NotificationJob) QueueName() string {
	return "notification-jobs"
}
