package internal

import (
	"crystalsage/internal/shards"
	"fmt"
	"io"
	"net/http"
)

type Crystal struct {
	Name   string         `yaml:"name"`
	Shards []shards.Shard `yaml:"shards"`
	Beam   chan string
	Auth   *AuthConfig
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
	mux.HandleFunc("/"+crystal.Name, Disperse)

	// Only register variant endpoint if at least one shard has a variant configured
	hasVariant := false
	for _, shard := range crystal.Shards {
		if shard.Variant != nil {
			hasVariant = true
			fmt.Printf("[Crystal][Register] Crystal '%s' has variant support, registering /%s/variant endpoint\n", crystal.Name, crystal.Name)
			break
		}
	}

	if hasVariant {
		mux.HandleFunc("/"+crystal.Name+"/variant", DisperseVariant)
	} else {
		fmt.Printf("[Crystal][Register] Crystal '%s' has no variants configured, skipping /%s/variant endpoint\n", crystal.Name, crystal.Name)
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

func DisperseVariant(w http.ResponseWriter, r *http.Request) {
	// Extract crystal name from path (remove /variant suffix)
	path := r.URL.Path[1:]
	crystalName := path[:len(path)-len("/variant")]
	fmt.Printf("[DisperseVariant][Start] Processing variant request for crystal: %s, Method: %s, Path: %s\n", crystalName, r.Method, r.URL.Path)

	crystal, exists := GlobalOrb.Crystals[crystalName]
	if !exists {
		fmt.Printf("[DisperseVariant][Error] Crystal '%s' not found\n", crystalName)
		http.Error(w, fmt.Sprintf("Crystal '%s' not found", crystalName), http.StatusNotFound)
		return
	}
	fmt.Printf("[DisperseVariant][Crystal] Found crystal '%s' with %d shards\n", crystalName, len(crystal.Shards))

	// Check authentication - priority: crystal-level > orb-level crystal > orb-level name
	fmt.Printf("[DisperseVariant][Auth] Checking authentication for crystal '%s'\n", crystalName)
	var auth *AuthConfig
	if crystal.Auth != nil {
		auth = crystal.Auth
		fmt.Printf("[DisperseVariant][Auth] Using crystal-level authentication\n")
	} else {
		auth = GetAuthForCrystal(crystalName, GlobalOrb)
		if auth != nil {
			fmt.Printf("[DisperseVariant][Auth] Using orb-level authentication\n")
		}
	}

	// If no crystal-level auth, check if any shard requires auth
	if auth == nil {
		fmt.Printf("[DisperseVariant][Auth] Checking shard-level authentication\n")
		for _, shard := range crystal.Shards {
			if shard.AuthKey != "" && GlobalOrb.Auth != nil {
				if shardAuth, ok := GlobalOrb.Auth.Shards[shard.AuthKey]; ok {
					auth = &shardAuth
					fmt.Printf("[DisperseVariant][Auth] Found shard-level auth for shard '%s'\n", shard.Alias)
					break // Use first shard auth found
				}
			}
		}
	}

	if auth != nil {
		fmt.Printf("[DisperseVariant][Auth] Authentication required, validating...\n")
		if !CheckAuth(w, r, auth) {
			fmt.Printf("[DisperseVariant][Auth] Authentication failed\n")
			return
		}
		fmt.Printf("[DisperseVariant][Auth] Authentication successful\n")
	} else {
		fmt.Printf("[DisperseVariant][Auth] No authentication required\n")
	}

	fmt.Printf("[DisperseVariant][Parse] Starting parameter extraction\n")
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("[DisperseVariant][Parse] Error parsing form: %v\n", err)
		crystal.Log("Unable to parse form", 0)
		return
	}
	fmt.Printf("[DisperseVariant][Parse] Form parsed successfully\n")

	// Get parameters from form or query
	fmt.Printf("[DisperseVariant][Parse] Extracting title, description, and color\n")
	title := r.FormValue("title")
	if title == "" {
		title = r.URL.Query().Get("title")
	}
	if title != "" {
		fmt.Printf("[DisperseVariant][Parse] Title: %s\n", title)
	}

	description := r.FormValue("description")
	if description == "" {
		description = r.URL.Query().Get("description")
	}
	if description != "" {
		fmt.Printf("[DisperseVariant][Parse] Description: %s\n", description)
	}

	colorStr := r.FormValue("color")
	if colorStr == "" {
		colorStr = r.URL.Query().Get("color")
	}
	if colorStr != "" {
		fmt.Printf("[DisperseVariant][Parse] Color: %s\n", colorStr)
	}

	// Default color if not provided
	color := 0
	if colorStr != "" {
		fmt.Sscanf(colorStr, "%d", &color)
		fmt.Printf("[DisperseVariant][Parse] Parsed color as integer: %d\n", color)
	}

	if title == "" && description == "" {
		fmt.Printf("[DisperseVariant][Error] Both title and description are empty\n")
		crystal.Log("Unable to get title or description", 0)
		return
	}

	// Send variant log to all shards that support it
	fmt.Printf("[DisperseVariant][Send] Sending variant log to shards (title: %s, description: %s, color: %d)\n", title, description, color)
	shardCount := 0
	for _, shard := range crystal.Shards {
		if shard.VariantLog != nil {
			fmt.Printf("[DisperseVariant][Send] Sending to shard '%s' (type: %s)\n", shard.Alias, shard.Type)
			shard.VariantLog(title, description, color)
			shardCount++
		} else {
			fmt.Printf("[DisperseVariant][Send] Skipping shard '%s' (no VariantLog function)\n", shard.Alias)
		}
	}
	fmt.Printf("[DisperseVariant][Complete] Sent variant log to %d shards\n", shardCount)
}
