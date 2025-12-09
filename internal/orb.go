package internal

import (
	"crystalsage/internal/shards"
	"fmt"
	"net/http"

	"github.com/gobuffalo/envy"
)

type OrbConfig struct {
	Global struct {
		Debug bool `yaml:"debug"`
		Port  int  `yaml:"port"`
	} `yaml:"global"`
	Variants map[string]shards.VariantConfig `yaml:"variants"`
	Auth     struct {
		Shards   map[string]AuthConfig `yaml:"shards"`
		Crystals map[string]AuthConfig `yaml:"crystals"`
		Names    map[string]AuthConfig `yaml:"names"`
	} `yaml:"auth"`
	Crystals []struct {
		Name   string `yaml:"name"`
		Shards []struct {
			Name    string `yaml:"name"`
			Crystal string `yaml:"crystal"`
			Alias   string `yaml:"alias"`
			Type    string `yaml:"type"`
			Variant string `yaml:"variant"`
			Auth    string `yaml:"auth"`
			EnvVar  bool   `yaml:"envVar"`
			Webhook string `yaml:"webhook"`
			ChatID  string `yaml:"chatId"`
		} `yaml:"shards"`
	} `yaml:"crystals"`
}

var GlobalOrb *Orb

type Orb struct {
	Port     int
	Debug    bool
	Crystals map[string]Crystal
	Auth     *struct {
		Shards   map[string]AuthConfig
		Crystals map[string]AuthConfig
		Names    map[string]AuthConfig
	}
}

func (orb *Orb) Load(orbConfig OrbConfig) {
	fmt.Println("[Orb][Loading]")
	fmt.Println("[Orb][Debug]", orbConfig.Global.Debug)
	fmt.Println("[Orb][Port]", orbConfig.Global.Port)
	orb.Debug = orbConfig.Global.Debug
	orb.Port = orbConfig.Global.Port
	orb.Crystals = make(map[string]Crystal)

	// Load authentication configuration
	if len(orbConfig.Auth.Shards) > 0 || len(orbConfig.Auth.Crystals) > 0 || len(orbConfig.Auth.Names) > 0 {
		orb.Auth = &struct {
			Shards   map[string]AuthConfig
			Crystals map[string]AuthConfig
			Names    map[string]AuthConfig
		}{
			Shards:   orbConfig.Auth.Shards,
			Crystals: orbConfig.Auth.Crystals,
			Names:    orbConfig.Auth.Names,
		}
		fmt.Println("[Orb][Auth] Authentication configured")
		if len(orbConfig.Auth.Shards) > 0 {
			fmt.Printf("[Orb][Auth] Shard-level auth: %d configured\n", len(orbConfig.Auth.Shards))
		}
		if len(orbConfig.Auth.Crystals) > 0 {
			fmt.Printf("[Orb][Auth] Crystal-level auth: %d configured\n", len(orbConfig.Auth.Crystals))
		}
		if len(orbConfig.Auth.Names) > 0 {
			fmt.Printf("[Orb][Auth] Name-level auth: %d configured\n", len(orbConfig.Auth.Names))
		}
	}
	for _, crystalCfg := range orbConfig.Crystals {
		crystal := Crystal{
			Name:   crystalCfg.Name,
			Shards: make([]shards.Shard, 0, len(crystalCfg.Shards)),
		}
		fmt.Println("[Orb][Crystal][Name]", crystalCfg.Name)

		// Load crystal-level authentication
		if orbConfig.Auth.Crystals != nil {
			if auth, ok := orbConfig.Auth.Crystals[crystalCfg.Name]; ok {
				crystal.Auth = &auth
				fmt.Printf("[Orb][Crystal][Auth] %s -> configured\n", crystalCfg.Name)
			}
		}
		// Also check names (same level, different key)
		if orbConfig.Auth.Names != nil {
			if auth, ok := orbConfig.Auth.Names[crystalCfg.Name]; ok {
				crystal.Auth = &auth
				fmt.Printf("[Orb][Crystal][Auth] %s -> configured via names\n", crystalCfg.Name)
			}
		}
		for _, shardCfg := range crystalCfg.Shards {
			fmt.Println("[Orb][Shard][Name]", shardCfg.Name)
			var shard shards.Shard

			// Load variant config if specified
			var variant *shards.VariantConfig
			if shardCfg.Variant != "" && orbConfig.Variants != nil {
				if v, ok := orbConfig.Variants[shardCfg.Variant]; ok {
					variant = &v
					fmt.Printf("[Orb][Shard][Variant] %s -> %s\n", shardCfg.Alias, shardCfg.Variant)
				} else {
					fmt.Printf("[Orb][Warning] Variant '%s' not found for shard: %s\n", shardCfg.Variant, shardCfg.Alias)
				}
			}

			// Load shard-level authentication if specified
			if shardCfg.Auth != "" && orbConfig.Auth.Shards != nil {
				if _, ok := orbConfig.Auth.Shards[shardCfg.Auth]; ok {
					fmt.Printf("[Orb][Shard][Auth] %s -> %s\n", shardCfg.Alias, shardCfg.Auth)
				} else {
					fmt.Printf("[Orb][Warning] Auth '%s' not found for shard: %s\n", shardCfg.Auth, shardCfg.Alias)
				}
			}

			shard = shards.Shard{
				EnvVar:  shardCfg.EnvVar,
				URL:     shardCfg.Webhook,
				Alias:   shardCfg.Alias,
				Debug:   orbConfig.Global.Debug,
				Variant: variant,
				AuthKey: shardCfg.Auth,
			}
			shard.Load()
			switch shardCfg.Type {
			case "slack":
				slack := shards.Slack{Shard: &shard}
				shard.Log = slack.Log
				shard.RawLog = slack.RawLog
				if variant != nil {
					shard.VariantLog = slack.VariantLog
				}
			case "discord":
				discord := shards.Discord{
					Shard: &shard,
				}
				shard.Log = discord.Log
				shard.RawLog = discord.RawLog
				if variant != nil {
					shard.VariantLog = discord.VariantLog
				}
			case "telegram":
				chatID := shardCfg.ChatID
				if shardCfg.EnvVar && chatID != "" {
					var err error
					chatID, err = envy.MustGet(chatID)
					if err != nil {
						fmt.Printf("[Orb][Warning] Failed to load chat_id from env for Telegram shard: %s\n", shardCfg.Alias)
					}
				}
				telegram := shards.Telegram{
					Shard:  &shard,
					ChatID: chatID,
				}
				shard.Log = telegram.Log
				shard.RawLog = telegram.RawLog
				if variant != nil {
					shard.VariantLog = telegram.VariantLog
				}
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
