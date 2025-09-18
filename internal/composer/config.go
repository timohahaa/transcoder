package composer

type (
	Config struct {
		PostgresDSN string `arg:"required,-,--,env:POSTGRES_DSN"`
		Redis
	}
	Redis struct {
		Addrs    []string `arg:"required,-,--,env:REDIS_ADDRS"`
		Username string   `arg:"required,-,--,env:REDIS_USERNAME"`
		Password string   `arg:"required,-,--,env:REDIS_PASSWORD"`
	}
)

func (c *Config) setDefaults() {}
