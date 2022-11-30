package config

type Configuration struct {
	Server     ServerConfig
	Onlyoffice OnlyofficeConfig
}

type ServerConfig struct {
	Host       string
	ServerPort string `json:"server_port"`
	Url        string `json:"url"`
}

type OnlyofficeConfig struct {
	Secret string
	Host   string
}
