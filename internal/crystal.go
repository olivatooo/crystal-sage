package internal

import (
	"crystalsage/internal/shards"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Crystal struct {
	Name     string                           `yaml:"name"`
	Shards   []shards.Shard                   `yaml:"shards"`
	Variants map[string]*shards.VariantConfig `yaml:"variants"`
	Beam     chan string
	Auth     *AuthConfig
}

func (crystal *Crystal) Log(content string, level uint8) {
	fmt.Printf("[Crystal][Log][Start] Crystal: %s, Level: %d, Shards: %d\n", crystal.Name, level, len(crystal.Shards))
	for i, shard := range crystal.Shards {
		fmt.Printf("[Crystal][Log] Processing shard %d/%d: %s (type: %s)\n", i+1, len(crystal.Shards), shard.Alias, shard.Type)
		shard.Log(content, level)
	}
	fmt.Printf("[Crystal][Log][Complete] Finished processing all shards for crystal '%s'\n", crystal.Name)
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
	// Always register base endpoint
	mux.HandleFunc("/"+crystal.Name, Disperse)
	fmt.Printf("[Crystal][Register] Registered base endpoint: /%s\n", crystal.Name)

	// Register variant endpoints by name
	if len(crystal.Variants) > 0 {
		for variantName := range crystal.Variants {
			endpoint := "/" + crystal.Name + "/" + variantName
			mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
				DisperseVariantByName(w, r, variantName)
			})
			fmt.Printf("[Crystal][Register] Registered variant endpoint: %s\n", endpoint)
		}
	} else {
		fmt.Printf("[Crystal][Register] Crystal '%s' has no variants configured\n", crystal.Name)
	}
}

func Disperse(w http.ResponseWriter, r *http.Request) {
	crystalName := r.URL.Path[1:]
	fmt.Printf("[Disperse][Start] Processing request for crystal: %s, Method: %s, Path: %s\n", crystalName, r.Method, r.URL.Path)

	crystal, exists := GlobalOrb.Crystals[crystalName]
	if !exists {
		fmt.Printf("[Disperse][Error] Crystal '%s' not found\n", crystalName)
		http.Error(w, fmt.Sprintf("Crystal '%s' not found", crystalName), http.StatusNotFound)
		return
	}
	fmt.Printf("[Disperse][Crystal] Found crystal '%s' with %d shards\n", crystalName, len(crystal.Shards))

	// Check authentication - priority: crystal-level > orb-level crystal > orb-level name
	fmt.Printf("[Disperse][Auth] Checking authentication for crystal '%s'\n", crystalName)
	var auth *AuthConfig
	if crystal.Auth != nil {
		auth = crystal.Auth
		fmt.Printf("[Disperse][Auth] Using crystal-level authentication\n")
	} else {
		auth = GetAuthForCrystal(crystalName, GlobalOrb)
		if auth != nil {
			fmt.Printf("[Disperse][Auth] Using orb-level authentication\n")
		}
	}

	// If no crystal-level auth, check if any shard requires auth
	if auth == nil {
		fmt.Printf("[Disperse][Auth] Checking shard-level authentication\n")
		for _, shard := range crystal.Shards {
			if shard.AuthKey != "" && GlobalOrb.Auth != nil {
				if shardAuth, ok := GlobalOrb.Auth.Shards[shard.AuthKey]; ok {
					auth = &shardAuth
					fmt.Printf("[Disperse][Auth] Found shard-level auth for shard '%s'\n", shard.Alias)
					break // Use first shard auth found
				}
			}
		}
	}

	if auth != nil {
		fmt.Printf("[Disperse][Auth] Authentication required, validating...\n")
		if !CheckAuth(w, r, auth) {
			fmt.Printf("[Disperse][Auth] Authentication failed\n")
			return
		}
		fmt.Printf("[Disperse][Auth] Authentication successful\n")
	} else {
		fmt.Printf("[Disperse][Auth] No authentication required\n")
	}

	fmt.Printf("[Disperse][Parse] Starting content extraction\n")
	var log string

	// Try to parse form data (works for application/x-www-form-urlencoded)
	fmt.Printf("[Disperse][Parse] Attempting to parse form data\n")
	err := r.ParseForm()
	if err == nil {
		log = r.FormValue("content")
		if log != "" {
			fmt.Printf("[Disperse][Parse] Found content in form data: %s\n", log)
		} else {
			fmt.Printf("[Disperse][Parse] No content found in form data\n")
		}
	} else {
		fmt.Printf("[Disperse][Parse] Error parsing form: %v\n", err)
	}

	// If not found in form, try query parameters
	if log == "" {
		fmt.Printf("[Disperse][Parse] Checking query parameters\n")
		log = r.URL.Query().Get("content")
		if log != "" {
			fmt.Printf("[Disperse][Parse] Found content in query parameters: %s\n", log)
		}
	}

	// If still not found and it's a POST/PUT with body, try reading raw body
	if log == "" && (r.Method == "POST" || r.Method == "PUT") && r.Body != nil {
		fmt.Printf("[Disperse][Parse] Attempting to read raw body (Method: %s)\n", r.Method)
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil && len(bodyBytes) > 0 {
			bodyStr := string(bodyBytes)
			fmt.Printf("[Disperse][Parse] Read %d bytes from body\n", len(bodyBytes))
			// Try to parse as simple key=value format (content=value)
			if len(bodyStr) > 8 && bodyStr[:8] == "content=" {
				log = bodyStr[8:]
				fmt.Printf("[Disperse][Parse] Extracted content from body (key=value format): %s\n", log)
			} else if len(bodyStr) > 0 {
				// Use entire body as content if it's plain text
				contentType := r.Header.Get("Content-Type")
				fmt.Printf("[Disperse][Parse] Content-Type: %s\n", contentType)
				if contentType == "" || contentType == "text/plain" {
					log = bodyStr
					fmt.Printf("[Disperse][Parse] Using entire body as content: %s\n", log)
				}
			}
		} else if err != nil {
			fmt.Printf("[Disperse][Parse] Error reading body: %v\n", err)
		}
	}

	if log == "" {
		fmt.Printf("[Disperse][Error] Unable to extract content from request\n")
		crystal.Log("Unable to get content", 0)
		return
	}

	fmt.Printf("[Disperse][Send] Sending content to %d shards: %s\n", len(crystal.Shards), log)
	crystal.Log(log, 0)
	fmt.Printf("[Disperse][Complete] Request processed successfully\n")
}

func DisperseVariantByName(w http.ResponseWriter, r *http.Request, variantName string) {
	// Extract crystal name from path
	// Path format: /crystal-name/variant-name
	path := r.URL.Path[1:] // Remove leading /
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		fmt.Printf("[DisperseVariantByName][Error] Invalid path format: %s\n", r.URL.Path)
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}
	crystalName := parts[0]
	fmt.Printf("[DisperseVariantByName][Start] Processing variant request for crystal: %s, variant: %s, Method: %s, Path: %s\n", crystalName, variantName, r.Method, r.URL.Path)

	crystal, exists := GlobalOrb.Crystals[crystalName]
	if !exists {
		fmt.Printf("[DisperseVariantByName][Error] Crystal '%s' not found\n", crystalName)
		http.Error(w, fmt.Sprintf("Crystal '%s' not found", crystalName), http.StatusNotFound)
		return
	}

	// Check if variant exists
	variant, variantExists := crystal.Variants[variantName]
	if !variantExists {
		fmt.Printf("[DisperseVariantByName][Error] Variant '%s' not found for crystal '%s'\n", variantName, crystalName)
		http.Error(w, fmt.Sprintf("Variant '%s' not found for crystal '%s'", variantName, crystalName), http.StatusNotFound)
		return
	}
	fmt.Printf("[DisperseVariantByName][Crystal] Found crystal '%s' with %d shards, variant '%s' found\n", crystalName, len(crystal.Shards), variantName)

	// Check authentication - priority: crystal-level > orb-level crystal > orb-level name
	fmt.Printf("[DisperseVariantByName][Auth] Checking authentication for crystal '%s'\n", crystalName)
	var auth *AuthConfig
	if crystal.Auth != nil {
		auth = crystal.Auth
		fmt.Printf("[DisperseVariantByName][Auth] Using crystal-level authentication\n")
	} else {
		auth = GetAuthForCrystal(crystalName, GlobalOrb)
		if auth != nil {
			fmt.Printf("[DisperseVariantByName][Auth] Using orb-level authentication\n")
		}
	}

	// If no crystal-level auth, check if any shard requires auth
	if auth == nil {
		fmt.Printf("[DisperseVariantByName][Auth] Checking shard-level authentication\n")
		for _, shard := range crystal.Shards {
			if shard.AuthKey != "" && GlobalOrb.Auth != nil {
				if shardAuth, ok := GlobalOrb.Auth.Shards[shard.AuthKey]; ok {
					auth = &shardAuth
					fmt.Printf("[DisperseVariantByName][Auth] Found shard-level auth for shard '%s'\n", shard.Alias)
					break // Use first shard auth found
				}
			}
		}
	}

	if auth != nil {
		fmt.Printf("[DisperseVariantByName][Auth] Authentication required, validating...\n")
		if !CheckAuth(w, r, auth) {
			fmt.Printf("[DisperseVariantByName][Auth] Authentication failed\n")
			return
		}
		fmt.Printf("[DisperseVariantByName][Auth] Authentication successful\n")
	} else {
		fmt.Printf("[DisperseVariantByName][Auth] No authentication required\n")
	}

	fmt.Printf("[DisperseVariantByName][Parse] Starting parameter extraction\n")
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("[DisperseVariantByName][Parse] Error parsing form: %v\n", err)
		crystal.Log("Unable to parse form", 0)
		return
	}
	fmt.Printf("[DisperseVariantByName][Parse] Form parsed successfully\n")

	// Get parameters from form or query
	fmt.Printf("[DisperseVariantByName][Parse] Extracting title, description, and color\n")
	title := r.FormValue("title")
	if title == "" {
		title = r.URL.Query().Get("title")
	}
	if title != "" {
		fmt.Printf("[DisperseVariantByName][Parse] Title: %s\n", title)
	}

	description := r.FormValue("description")
	if description == "" {
		description = r.URL.Query().Get("description")
	}
	if description != "" {
		fmt.Printf("[DisperseVariantByName][Parse] Description: %s\n", description)
	}

	colorStr := r.FormValue("color")
	if colorStr == "" {
		colorStr = r.URL.Query().Get("color")
	}
	if colorStr != "" {
		fmt.Printf("[DisperseVariantByName][Parse] Color: %s\n", colorStr)
	}

	// Default color if not provided
	color := 0
	if colorStr != "" {
		fmt.Sscanf(colorStr, "%d", &color)
		fmt.Printf("[DisperseVariantByName][Parse] Parsed color as integer: %d\n", color)
	}

	if title == "" && description == "" {
		fmt.Printf("[DisperseVariantByName][Error] Both title and description are empty\n")
		crystal.Log("Unable to get title or description", 0)
		return
	}

	// Send variant log to all shards using the variant config
	fmt.Printf("[DisperseVariantByName][Send] Sending variant log to shards (variant: %s, title: %s, description: %s, color: %d)\n", variantName, title, description, color)
	shardCount := 0
	for _, shard := range crystal.Shards {
		if shard.VariantLog != nil {
			fmt.Printf("[DisperseVariantByName][Send] Sending to shard '%s' (type: %s) with variant '%s'\n", shard.Alias, shard.Type, variantName)
			shard.VariantLog(variant, title, description, color)
			shardCount++
		} else {
			fmt.Printf("[DisperseVariantByName][Send] Skipping shard '%s' (no VariantLog function)\n", shard.Alias)
		}
	}
	fmt.Printf("[DisperseVariantByName][Complete] Sent variant log to %d shards\n", shardCount)
}
