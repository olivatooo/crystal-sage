package internal

import (
	"crystalsage/internal/shards"
	"encoding/json"
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
		Name     string   `yaml:"name"`
		Variants []string `yaml:"variants"`
		Shards   []struct {
			Name    string `yaml:"name"`
			Crystal string `yaml:"crystal"`
			Alias   string `yaml:"alias"`
			Type    string `yaml:"type"`
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
			Name:     crystalCfg.Name,
			Shards:   make([]shards.Shard, 0, len(crystalCfg.Shards)),
			Variants: make(map[string]*shards.VariantConfig),
		}
		fmt.Println("[Orb][Crystal][Name]", crystalCfg.Name)

		// Load variants for this crystal
		if len(crystalCfg.Variants) > 0 && orbConfig.Variants != nil {
			fmt.Printf("[Orb][Crystal][Variants] Loading %d variants for crystal '%s'\n", len(crystalCfg.Variants), crystalCfg.Name)
			for _, variantName := range crystalCfg.Variants {
				if variant, ok := orbConfig.Variants[variantName]; ok {
					crystal.Variants[variantName] = &variant
					fmt.Printf("[Orb][Crystal][Variants] Loaded variant '%s' for crystal '%s'\n", variantName, crystalCfg.Name)
				} else {
					fmt.Printf("[Orb][Warning] Variant '%s' not found in variants config for crystal '%s'\n", variantName, crystalCfg.Name)
				}
			}
		}

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
				Type:    shardCfg.Type,
				Debug:   orbConfig.Global.Debug,
				Variant: nil, // Variants are now at crystal level
				AuthKey: shardCfg.Auth,
			}
			shard.Load()
			switch shardCfg.Type {
			case "slack":
				slack := shards.Slack{Shard: &shard}
				shard.Log = slack.Log
				shard.RawLog = slack.RawLog
				// Always enable VariantLog if crystal has variants
				if len(crystal.Variants) > 0 {
					shard.VariantLog = slack.VariantLog
				}
			case "discord":
				discord := shards.Discord{
					Shard: &shard,
				}
				shard.Log = discord.Log
				shard.RawLog = discord.RawLog
				// Always enable VariantLog if crystal has variants
				if len(crystal.Variants) > 0 {
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
				// Always enable VariantLog if crystal has variants
				if len(crystal.Variants) > 0 {
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
	// Register root endpoints
	mux.HandleFunc("/", RootHandler)
	for _, crystal := range orb.Crystals {
		crystal.Register(mux)
	}
	fmt.Println("[Orb][Registered]")
}

// RootHandler handles requests to the root path
func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[RootHandler][Start] Processing request: Method=%s, Path=%s\n", r.Method, r.URL.Path)

	if r.Method == http.MethodHead {
		// Health check - just return 200
		fmt.Printf("[RootHandler] Health check request, returning 200 OK\n")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		fmt.Printf("[RootHandler] GET request, building services list\n")
		// Return all available log services
		services := make([]ServiceInfo, 0, len(GlobalOrb.Crystals))
		for name, crystal := range GlobalOrb.Crystals {
			shardTypes := make(map[string]int)
			hasAuth := crystal.Auth != nil

			// Check if any shard requires auth
			if !hasAuth && GlobalOrb.Auth != nil {
				for _, shard := range crystal.Shards {
					if shard.AuthKey != "" {
						if _, ok := GlobalOrb.Auth.Shards[shard.AuthKey]; ok {
							hasAuth = true
							break
						}
					}
				}
			}

			// Also check crystal-level auth
			if !hasAuth {
				hasAuth = GetAuthForCrystal(name, GlobalOrb) != nil
			}

			// Count shard types
			for _, shard := range crystal.Shards {
				shardType := shard.Type
				if shardType == "" {
					shardType = "unknown"
				}
				shardTypes[shardType]++
			}

			// Build endpoints list - always include base endpoint
			endpoints := []string{
				fmt.Sprintf("/%s", name),
			}

			// Add variant endpoints by name
			if len(crystal.Variants) > 0 {
				for variantName := range crystal.Variants {
					endpoints = append(endpoints, fmt.Sprintf("/%s/%s", name, variantName))
				}
			}

			service := ServiceInfo{
				Name:         name,
				Shards:       len(crystal.Shards),
				ShardTypes:   shardTypes,
				RequiresAuth: hasAuth,
				Endpoints:    endpoints,
			}
			services = append(services, service)
		}

		response := ServicesResponse{
			Services: services,
			Count:    len(services),
		}

		fmt.Printf("[RootHandler] Returning %d services\n", len(services))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		fmt.Printf("[RootHandler][Complete] Services list returned successfully\n")
		return
	}

	// Method not allowed
	fmt.Printf("[RootHandler][Error] Method not allowed: %s\n", r.Method)
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// ServiceInfo represents information about a log service (crystal)
type ServiceInfo struct {
	Name         string         `json:"name"`
	Shards       int            `json:"shards"`
	ShardTypes   map[string]int `json:"shard_types"`
	RequiresAuth bool           `json:"requires_auth"`
	Endpoints    []string       `json:"endpoints"`
}

// ServicesResponse represents the response containing all services
type ServicesResponse struct {
	Services []ServiceInfo `json:"services"`
	Count    int           `json:"count"`
}
