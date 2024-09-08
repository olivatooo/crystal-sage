package internal

import (
	"crystalsage/internal/shards"
	"fmt"
	"net/http"
)

type OrbConfig struct {
	Global struct {
		Debug bool `yaml:"debug"`
		Port  int  `yaml:"port"`
	} `yaml:"global"`
	Crystals []struct {
		Name   string `yaml:"name"`
		Shards []struct {
			Name    string `yaml:"name"`
			Crystal string `yaml:"crystal"`
			Alias   string `yaml:"alias"`
			Type    string `yaml:"type"`
			EnvVar  bool   `yaml:"envVar"`
			Webhook string `yaml:"webhook"`
		} `yaml:"shards"`
	} `yaml:"crystals"`
}

var GlobalOrb *Orb

type Orb struct {
	Port     int
	Debug    bool
	Crystals map[string]Crystal
}

func (orb *Orb) Load(orbConfig OrbConfig) {
	fmt.Println("[Orb][Loading]")
	fmt.Println("[Orb][Debug]", orbConfig.Global.Debug)
	fmt.Println("[Orb][Port]", orbConfig.Global.Port)
	orb.Debug = orbConfig.Global.Debug
	orb.Port = orbConfig.Global.Port
	orb.Crystals = make(map[string]Crystal)
	for _, crystalCfg := range orbConfig.Crystals {
		crystal := Crystal{
			Name:   crystalCfg.Name,
			Shards: make([]shards.Shard, 0, len(crystalCfg.Shards)),
		}
		fmt.Println("[Orb][Crystal][Name]", crystalCfg.Name)
		for _, shardCfg := range crystalCfg.Shards {
			fmt.Println("[Orb][Shard][Name]", shardCfg.Name)
			var shard shards.Shard
			shard = shards.Shard{
				EnvVar: shardCfg.EnvVar,
				URL:    shardCfg.Webhook,
				Alias:  shardCfg.Alias,
				Debug:  orbConfig.Global.Debug,
			}
			shard.Load()
			switch shardCfg.Type {
			case "slack":
				slack := shards.Slack{Shard: &shard}
				shard.Log = slack.Log
				shard.RawLog = slack.RawLog
			case "discord":
				discord := shards.Discord{
					Shard: &shard,
				}
				shard.Log = discord.Log
				shard.RawLog = discord.RawLog
			default:
				fmt.Printf("[Orb][Warning] Unknown shard type: %s\n", shardCfg.Type)
				continue
			}
			crystal.AppendShard(shard)
		}
		orb.Crystals[crystalCfg.Name] = crystal
	}
	fmt.Println("[Orb][Loaded]")
}

func (orb *Orb) Register(mux *http.ServeMux) {
	fmt.Println("[Orb][Registering]")
	for _, crystal := range orb.Crystals {
		crystal.Register(mux)
	}
	fmt.Println("[Orb][Registered]")
}
