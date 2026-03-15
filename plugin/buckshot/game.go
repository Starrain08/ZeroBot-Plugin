package buckshot

import (
	"sync"
)

type Player struct {
	name     string
	id       int64
	hp       int
	items    []string
	handcuff bool
}

type Game struct {
	player1          *Player
	player2          *Player
	status           string
	bullet           []string
	currentTurn      int
	double           bool
	round            int
	usedHandcuff     bool
	waitingForItem   bool
	waitingPlayer    int
	waitingForItemMu sync.Mutex
	mu               sync.Mutex
}

type MuteConfig struct {
	enabled  bool
	duration int
}

var (
	games               = make(map[int64]*Game)
	gamesMutex          sync.RWMutex
	dontDispose         = make(map[int64]func())
	epinephrineTimeouts = make(map[int64]chan struct{})
	muteConfig          = MuteConfig{enabled: false, duration: 0}
	muteConfigMutex     sync.RWMutex
)

func (p *Player) removeItem(item string) {
	for i, v := range p.items {
		if v == item {
			p.items = append(p.items[:i], p.items[i+1:]...)
			break
		}
	}
}

func (g *Game) getPlayer(player int) *Player {
	if player == 1 {
		return g.player1
	}
	return g.player2
}

func getGame(gid int64) (*Game, bool) {
	gamesMutex.RLock()
	defer gamesMutex.RUnlock()
	game, exists := games[gid]
	return game, exists
}

func setGame(gid int64, game *Game) {
	gamesMutex.Lock()
	defer gamesMutex.Unlock()
	games[gid] = game
}

func deleteGame(gid int64) {
	gamesMutex.Lock()
	defer gamesMutex.Unlock()
	delete(games, gid)
}

func getChannelIDFromEvent(groupID, userID int64) int64 {
	if groupID == 0 {
		return -userID
	}
	return groupID
}

func getMuteConfig() MuteConfig {
	muteConfigMutex.RLock()
	defer muteConfigMutex.RUnlock()
	return muteConfig
}

func setMuteConfig(enabled bool, duration int) {
	muteConfigMutex.Lock()
	defer muteConfigMutex.Unlock()
	muteConfig.enabled = enabled
	muteConfig.duration = duration
}
