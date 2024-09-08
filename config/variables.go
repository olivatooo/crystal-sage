package config

import "github.com/gobuffalo/envy"

type env struct {
	YAML_PATH string
}

var ENV env

func LoadEnvironment() {
	envy.Load(".env")
	ENV.YAML_PATH = envy.Get("YAML_PATH", "config.yaml")
}
