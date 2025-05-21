package config

type Config struct {
	ServerPort string
	MapWidth   int
	MapHeight  int
}

func LoadConfig() (*Config, error) {
	return &Config{
		ServerPort: "8080",
		MapWidth:   20,
		MapHeight:  20,
	}, nil
}
