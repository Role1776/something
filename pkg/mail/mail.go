package mail

type MailData struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Send(input *MailData) error
}
