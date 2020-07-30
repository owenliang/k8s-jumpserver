package config

import (
	"github.com/BurntSushi/toml"
)

type Server struct {
	Listen string `toml:"listen"`
}

type Record struct {
	Path string `toml:"path"`
}

type JumpServer struct {
	Server *Server	`toml:"Server"`
	Record *Record `toml:"Record"`
}

var G_JumpServer = &JumpServer{}

func LoadConfig(filename *string) (err error) {
	if _, err = toml.DecodeFile(*filename, G_JumpServer); err != nil {
		return
	}
	return
}
