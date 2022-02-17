package main

type Config struct {
	Commands []Command `yaml:"commands"`
}

type Command struct {
	Name         string `yaml:"name"`
	Cmd          string `yaml:"cmd"`
	SleepSeconds int    `yaml:"sleep,omitempty"`
	ShowLog      bool   `yaml:"showLog,omitempty"`
}
