package manager

type ServerConfig struct {
	Network string `config:"network"`
	Address string `config:"address"`
	WorkDir string `config:"work.dir"`
}

var DefaultServerConfig = ServerConfig{
	Network: "tcp",
	Address: ":30398",
	WorkDir: "/tmp",
}
