package model

// LoginForm holds raw admin login form input.
type LoginForm struct {
	Email    string
	Password string
	Errors   map[string]string
}

func NewLoginForm() LoginForm {
	return LoginForm{
		Errors: make(map[string]string),
	}
}

func (f LoginForm) HasErrors() bool {
	return len(f.Errors) > 0
}
