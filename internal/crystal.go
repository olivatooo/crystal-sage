package internal

import (
	"crystalsage/internal/shards"
	"fmt"
	"net/http"
)

type Crystal struct {
	Name   string         `yaml:"name"`
	Shards []shards.Shard `yaml:"shards"`
	Beam   chan string
	Auth   *AuthConfig
}

func (crystal *Crystal) Log(content string, level uint8) {
	for _, shard := range crystal.Shards {
		shard.Log(content, level)
	}
}

func (crystal *Crystal) RawLog(content string) {
	for _, shard := range crystal.Shards {
		shard.RawLog(content)
	}
}

func (crystal *Crystal) AppendShard(shard shards.Shard) {
	crystal.Shards = append(crystal.Shards, shard)
}

func (crystal *Crystal) Register(mux *http.ServeMux) {
	mux.HandleFunc("/"+crystal.Name, Disperse)
	mux.HandleFunc("/"+crystal.Name+"/variant", DisperseVariant)
}

func Disperse(w http.ResponseWriter, r *http.Request) {
	crystalName := r.URL.Path[1:]
	crystal := GlobalOrb.Crystals[crystalName]

	// Check authentication - priority: crystal-level > orb-level crystal > orb-level name
	var auth *AuthConfig
	if crystal.Auth != nil {
		auth = crystal.Auth
	} else {
		auth = GetAuthForCrystal(crystalName, GlobalOrb)
	}

	// If no crystal-level auth, check if any shard requires auth
	if auth == nil {
		for _, shard := range crystal.Shards {
			if shard.AuthKey != "" && GlobalOrb.Auth != nil {
				if shardAuth, ok := GlobalOrb.Auth.Shards[shard.AuthKey]; ok {
					auth = &shardAuth
					break // Use first shard auth found
				}
			}
		}
	}

	if auth != nil {
		if !CheckAuth(w, r, auth) {
			return
		}
	}

	err := r.ParseForm()
	if err != nil {
		crystal.Log("Unable to parse form", 0)
		return
	}
	log := r.FormValue("content")
	if log == "" {
		query := r.URL.Query()
		log = query.Get("content")
	}
	if log == "" {
		crystal.Log("Unable to get content", 0)
		return
	}
	crystal.Log(log, 0)
}

func DisperseVariant(w http.ResponseWriter, r *http.Request) {
	// Extract crystal name from path (remove /variant suffix)
	path := r.URL.Path[1:]
	crystalName := path[:len(path)-len("/variant")]
	crystal := GlobalOrb.Crystals[crystalName]

	// Check authentication - priority: crystal-level > orb-level crystal > orb-level name
	var auth *AuthConfig
	if crystal.Auth != nil {
		auth = crystal.Auth
	} else {
		auth = GetAuthForCrystal(crystalName, GlobalOrb)
	}

	// If no crystal-level auth, check if any shard requires auth
	if auth == nil {
		for _, shard := range crystal.Shards {
			if shard.AuthKey != "" && GlobalOrb.Auth != nil {
				if shardAuth, ok := GlobalOrb.Auth.Shards[shard.AuthKey]; ok {
					auth = &shardAuth
					break // Use first shard auth found
				}
			}
		}
	}

	if auth != nil {
		if !CheckAuth(w, r, auth) {
			return
		}
	}

	err := r.ParseForm()
	if err != nil {
		crystal.Log("Unable to parse form", 0)
		return
	}

	// Get parameters from form or query
	title := r.FormValue("title")
	if title == "" {
		title = r.URL.Query().Get("title")
	}

	description := r.FormValue("description")
	if description == "" {
		description = r.URL.Query().Get("description")
	}

	colorStr := r.FormValue("color")
	if colorStr == "" {
		colorStr = r.URL.Query().Get("color")
	}

	// Default color if not provided
	color := 0
	if colorStr != "" {
		fmt.Sscanf(colorStr, "%d", &color)
	}

	if title == "" && description == "" {
		crystal.Log("Unable to get title or description", 0)
		return
	}

	// Send variant log to all shards that support it
	for _, shard := range crystal.Shards {
		if shard.VariantLog != nil {
			shard.VariantLog(title, description, color)
		}
	}
}
