package internal

import "sync"

// Only used if we want concurrent safe logs
var Mu sync.Mutex
