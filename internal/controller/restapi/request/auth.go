package request

type Register struct {
	Token string
	Login string
	Pswd  string
}

type Auth struct {
	Login string
	Pswd  string
}
