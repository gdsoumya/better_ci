package types

type EventDetails struct {
	Permission string `json:"author_association"`
	Body       string `json:"body"`
	URL        string `json:"html_url"`
	ID         int64  `json:"id"`
	Username   string
	Repo       string
	PR         string
	Time       int
}
