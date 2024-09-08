package internal

import (
	"crystalsage/internal/shards"
	"net/http"
)

type Crystal struct {
	Name   string         `yaml:"name"`
	Shards []shards.Shard `yaml:"shards"`
	Beam   chan string
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
}

func Disperse(w http.ResponseWriter, r *http.Request) {
	crystal := GlobalOrb.Crystals[r.URL.Path[1:]]
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
