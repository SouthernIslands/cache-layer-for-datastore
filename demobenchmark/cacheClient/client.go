package cacheClient

type Cmd struct {
	Name  string
	Key   string
	Value BookData
	Error error
	Res   bool
}

type BookData struct {
	ISBN       string
	BookTitle  string
	BookAuthor string
	Year       string
	Publisher  string
	ImageM     string
}

type Book struct {
	ID            int64
	Title         string
	Author        string
	PublishedDate string
	ImageURL      string
	Description   string
	CreatedBy     string
	CreatedByID   string
}

type Client interface {
	Run(*Cmd)
}

func New(server string) Client {
	return newHTTPClient(server)
}
