package shards

import (
	"fmt"

	"github.com/gobuffalo/envy"
)

type Shard struct {
	URL        string
	Name       string
	Alias      string
	Type       string // Shard type: discord, slack, telegram
	Debug      bool
	Log        func(string, uint8)
	EnvVar     bool
	RawLog     func(string)
	Variant    *VariantConfig
	VariantLog func(string, string, int) // title, description, color
	AuthKey    string                    // Reference to auth config name
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
