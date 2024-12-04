package main

type Config struct {
	Remote    string
	ApiKey    string
	ApiSecret string
	Verbose   bool
	NoPrompt  bool
}

func (cf *Config) ApplyDefaults() {
	cf.Remote = "https://openqa.opensuse.org"
	cf.ApiKey = ""
	cf.ApiSecret = ""
	cf.Verbose = false
	cf.NoPrompt = false
}
