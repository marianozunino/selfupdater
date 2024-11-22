package selfupdater

type Option func(*Updater)

func WithToken(token string) Option {
	return func(c *Updater) {
		c.client = c.client.WithAuthToken(token)
	}
}

type Asset struct {
	ID       int64
	Name     string
	Checksum string
}
