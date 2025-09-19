package composer

type (
	Config struct {
		PostgresDSN string `arg:"required,-,--,env:POSTGRES_DSN"`
		HttpAddr    string `arg:"required,-,--,env:HTTP_ADDR"`
		Redis
		Splitter
	}
	Redis struct {
		Addrs    []string `arg:"required,-,--,env:REDIS_ADDRS"`
		Username string   `arg:"required,-,--,env:REDIS_USERNAME"`
		Password string   `arg:"required,-,--,env:REDIS_PASSWORD"`
	}
	Splitter struct {
		Workers  int `arg:"-,--,env:SPLITTER_WORKERS"`
		Watchers int `arg:"-,--,env:SPLITTER_WATCHERS"`
	}
)

func (c *Config) setDefaults() {
	if c.Splitter.Workers <= 0 {
		c.Splitter.Workers = 5
	}
	if c.Splitter.Watchers <= 0 {
		c.Splitter.Watchers = 1
	}
}
