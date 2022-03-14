package quark

type AuthenticateFunc func(c *Console) bool

type Option struct {
	Authenticate AuthenticateFunc
	PathPrefix   []string
}

func (q *Quark) WithAuthenticate(f AuthenticateFunc) {
	q.option.Authenticate = f
}

func (q *Quark) WithPathPrefix(p []string) {
	q.option.PathPrefix = p
}
