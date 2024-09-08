package shards

import (
	"fmt"

	"github.com/gobuffalo/envy"
)

type Shard struct {
	URL    string
	Name   string
	Alias  string
	Debug  bool
	Log    func(string, uint8)
	EnvVar bool
	RawLog func(string)
}

func (shard *Shard) Load() {
	fmt.Println("[Shard][Loading]", shard.Alias)
	if !shard.EnvVar {
		return
	}
	var err error
	shard.URL, err = envy.MustGet(shard.URL)
	if err != nil {
		return
	}
}
