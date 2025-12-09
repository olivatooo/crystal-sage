package internal

import (
	"fmt"
	"net/http"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Key string `yaml:"key"`
}

// CheckAuth validates the X-Api-Key header against the provided auth config
func CheckAuth(w http.ResponseWriter, r *http.Request, auth *AuthConfig) bool {
	fmt.Printf("[CheckAuth][Start] Validating authentication\n")
	if auth == nil || auth.Key == "" {
		// No authentication required
		fmt.Printf("[CheckAuth] No authentication required (auth is nil or key is empty)\n")
		return true
	}

	fmt.Printf("[CheckAuth] Authentication required, checking X-Api-Key header\n")
	apiKey := r.Header.Get("X-Api-Key")
	if apiKey == "" {
		fmt.Printf("[CheckAuth][Error] Missing X-Api-Key header\n")
		http.Error(w, "Missing X-Api-Key header", http.StatusUnauthorized)
		return false
	}

	fmt.Printf("[CheckAuth] X-Api-Key header present, validating key\n")
	if apiKey != auth.Key {
		fmt.Printf("[CheckAuth][Error] Invalid API key (provided key does not match)\n")
		http.Error(w, "Invalid API key", http.StatusForbidden)
		return false
	}

	fmt.Printf("[CheckAuth][Success] API key validated successfully\n")
	return true
}

// GetAuthForShard retrieves authentication config for a shard
// Priority: shard-specific > crystal-specific > crystal-name-specific
func GetAuthForShard(crystalName string, shardAlias string, orb *Orb) *AuthConfig {
	if orb.Auth == nil {
		return nil
	}

	// Check shard-specific auth by alias
	if auth, ok := orb.Auth.Shards[shardAlias]; ok {
		return &auth
	}

	// Check crystal-specific auth
	if auth, ok := orb.Auth.Crystals[crystalName]; ok {
		return &auth
	}

	// Check by crystal name
	if auth, ok := orb.Auth.Names[crystalName]; ok {
		return &auth
	}

	return nil
}

// GetAuthForCrystal retrieves authentication config for a crystal
func GetAuthForCrystal(crystalName string, orb *Orb) *AuthConfig {
	// Check crystal-specific auth
	if orb.Auth != nil {
		if auth, ok := orb.Auth.Crystals[crystalName]; ok {
			return &auth
		}
		// Check by name
		if auth, ok := orb.Auth.Names[crystalName]; ok {
			return &auth
		}
	}

	return nil
}
