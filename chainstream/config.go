package chainstream

type Config struct {
	WssApiEndpoint string
}

func NewConfig(wssApiEndpoint string) *Config {
	return &Config{
		WssApiEndpoint: wssApiEndpoint,
	}
}
