package main

type Config struct {
	Version   string
	LastBlock int32
}

func (c *Config) GetVersion() string {
	return c.Version
}

func (c *Config) GetLastBlock() int32 {
	return c.LastBlock
}
