package query

import (
	// comp "cinco-paus/component"
	comp "cinco-paus/component"
	"fmt"
	"log"

	"pkg.world.dev/world-engine/cardinal"
	"pkg.world.dev/world-engine/cardinal/search/filter"
	"pkg.world.dev/world-engine/cardinal/types"
)

type PlayerData struct {
	ID         types.EntityID `json:"id"`
	X          int            `json:"x"`
	Y          int            `json:"y"`
	MaxHealth  int            `json:"maxHealth"`
	CurrHealth int            `json:"currHealth"`
}

type WandData struct {
	ID          types.EntityID `json:"id"`
	Number      int            `json:"number"`
	IsAvailable bool           `json:"isAvailable"`
}

type WallData struct {
	ID   types.EntityID `json:"id"`
	X    int            `json:"x"`
	Y    int            `json:"y"`
	Type int            `json:"type"`
}

type MonsterData struct {
	ID         types.EntityID `json:"id"`
	X          int            `json:"x"`
	Y          int            `json:"y"`
	CurrHealth int            `json:"currHealth"`
	MaxHealth  int            `json:"maxHealth"`
	Type       int            `json:"type"`
}

type GameStateRequest struct {
	GameID types.EntityID
}

type GameStateResponse struct {
	GameID   types.EntityID `json:"gameID"`
	Level    int            `json:"level"`
	Score    int            `json:"score"`
	Player   PlayerData     `json:"player"`
	Reveals  [][]int        `json:"reveals"`
	Wands    []WandData     `json:"wands"`
	Walls    []WallData     `json:"walls"`
	Monsters []MonsterData  `json:"monsters"`
}

func GameState(world cardinal.WorldContext, req *GameStateRequest) (*GameStateResponse, error) {
	playerData := &PlayerData{}
	wands := &[]WandData{}
	walls := &[]WallData{}
	monsters := &[]MonsterData{}
	gameID := req.GameID

	var outsideErr error

	searchErr := cardinal.NewSearch(
		world,
		filter.Contains(comp.GameObj{})).
		Each(func(id types.EntityID) bool {
			gameObj, err := cardinal.GetComponent[comp.GameObj](world, id)
			if err != nil {
				log.Println("gameObj err: ", err)
				return false
			}
			// if object is from the wrong game, skip to next object
			if gameObj.GameID != gameID {
				return true
			}

			outsideErr = getPlayerData(world, id, playerData)
			if outsideErr != nil {
				log.Printf("error getting player data: %v\n", outsideErr)
				return false
			}

			outsideErr = getWandData(world, id, wands)
			if outsideErr != nil {
				log.Printf("error getting wand data: %v\n", outsideErr)
				return false
			}

			outsideErr = getWallData(world, id, walls)
			if outsideErr != nil {
				log.Printf("error getting wall data: %v\n", outsideErr)
				return false
			}

			outsideErr = getMonsterData(world, id, monsters)
			if outsideErr != nil {
				log.Printf("error getting monster data: %v\n", outsideErr)
				return false
			}

			// always check next entity
			return true
		})

	if searchErr != nil {
		log.Printf("searchErr: %v\n", searchErr)
		return nil, searchErr
	}
	if outsideErr != nil {
		log.Printf("outsideErr: %v\n", outsideErr)
		return nil, outsideErr
	}

	game, err := cardinal.GetComponent[comp.Game](world, gameID)
	if err != nil {
		return nil, err
	}

	return &GameStateResponse{
		GameID:   gameID,
		Level:    game.Level,
		Score:    game.Score,
		Reveals:  *game.Reveals,
		Player:   *playerData,
		Wands:    *wands,
		Walls:    *walls,
		Monsters: *monsters,
	}, nil
}

func getPlayerData(world cardinal.WorldContext, id types.EntityID, playerData *PlayerData) error {
	player, _ := cardinal.GetComponent[comp.Player](world, id) // don't error check, want to ignore unfound errors
	if player != nil {
		pos, err := cardinal.GetComponent[comp.Position](world, id)
		if err != nil {
			return fmt.Errorf("failed to get position component for player: %w", err)
		}
		health, err := cardinal.GetComponent[comp.Health](world, id)
		if err != nil {
			return fmt.Errorf("failed to get position component for health: %w", err)
		}
		playerData.ID = id
		playerData.X = pos.X
		playerData.Y = pos.Y
		playerData.MaxHealth = health.MaxHealth
		playerData.CurrHealth = health.CurrHealth
	}
	return nil
}

func getWandData(world cardinal.WorldContext, id types.EntityID, wands *[]WandData) error {
	wand, _ := cardinal.GetComponent[comp.WandCore](world, id)
	if wand != nil {
		availableObj, err := cardinal.GetComponent[comp.Available](world, id)
		if err != nil {
			return fmt.Errorf("failed to get available component for wand: %w", err)
		}
		*wands = append(*wands, WandData{
			ID:          id,
			Number:      wand.Number,
			IsAvailable: availableObj.IsAvailable,
		})
	}
	return nil
}

func getWallData(world cardinal.WorldContext, id types.EntityID, walls *[]WallData) error {
	wall, _ := cardinal.GetComponent[comp.Wall](world, id)
	if wall != nil {
		pos, err := cardinal.GetComponent[comp.Position](world, id)
		if err != nil {
			return fmt.Errorf("failed to get position component for wall: %w", err)
		}
		*walls = append(*walls, WallData{
			ID:   id,
			X:    pos.X,
			Y:    pos.Y,
			Type: int(comp.WallCollide),
		})
	}
	return nil
}

func getMonsterData(world cardinal.WorldContext, id types.EntityID, monsters *[]MonsterData) error {
	monster, _ := cardinal.GetComponent[comp.Monster](world, id)
	if monster != nil {
		pos, err := cardinal.GetComponent[comp.Position](world, id)
		if err != nil {
			return fmt.Errorf("failed to get position component for monster: %w", err)
		}
		health, err := cardinal.GetComponent[comp.Health](world, id)
		if err != nil {
			return fmt.Errorf("failed to get position component for monster: %w", err)
		}
		*monsters = append(*monsters, MonsterData{
			ID:         id,
			X:          pos.X,
			Y:          pos.Y,
			CurrHealth: health.CurrHealth,
			MaxHealth:  health.MaxHealth,
			Type:       int(comp.MonsterCollide),
		})
	}
	return nil
}
