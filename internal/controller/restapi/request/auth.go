package request

type Register struct {
	AdminToken string
	Login      string
	Password   string
}

type Auth struct {
	Login    string
	Password string
}
