package encoder

type (
	Config struct {
		ComposerAddrs []string `arg:"required,-,--,env:COMPOSER_ADDRS"`
	}
)

func (c *Config) setDefaults() {}
