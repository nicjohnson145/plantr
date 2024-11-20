package vault

import (
	"github.com/rs/zerolog"
)

type HashicorpConfig struct {
	Logger zerolog.Logger
}

func NewHashicorp(conf HashicorpConfig) *Hashicorp {
	return &Hashicorp{
		log: conf.Logger,
	}
}

type Hashicorp struct {
	log zerolog.Logger
}


func (h *Hashicorp) GetSecretVersion() (string, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Hashicorp) ReadSecretData() (map[string]any, error) {
	panic("not implemented") // TODO: Implement
}

