package response

type Register struct {
	Login string `json:"login"`
}

type Auth struct {
	Token string `json:"token"`
}

type Logout map[string]bool
