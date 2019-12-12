package model

// Server ..
type Server struct {
	Host         string
	User         string
	Password     string
	IdentityFile string
	Port         string

	Lable string
	Tags  []string
}
