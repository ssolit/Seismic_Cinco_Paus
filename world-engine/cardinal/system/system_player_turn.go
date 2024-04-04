package system

import (
	comp "cinco-paus/component"
	"cinco-paus/msg"
	"cinco-paus/seismic/client"
	"errors"
	"fmt"
	"strconv"

	"pkg.world.dev/world-engine/cardinal"
	"pkg.world.dev/world-engine/cardinal/message"
	"pkg.world.dev/world-engine/cardinal/types"
)

func PlayerTurnSystem(world cardinal.WorldContext) error {
	return cardinal.EachMessage[msg.PlayerTurnMsg, msg.PlayerTurnResult](
		world,
		func(turn message.TxData[msg.PlayerTurnMsg]) (msg.PlayerTurnResult, error) {
			fmt.Println("starting player turn system")
			err := turn.Msg.ValFmt()
			if err != nil {
				return msg.PlayerTurnResult{}, fmt.Errorf("error with msg format: %w", err)
			}

			direction, err := comp.StringToDirection(turn.Msg.Direction)
			if err != nil {
				return msg.PlayerTurnResult{}, err
			}

			eventLogList := &[]comp.GameEventLog{}

			switch turn.Msg.Action {
			case "attack":
				err = player_turn_attack(world, direction)
				if err != nil {
					return msg.PlayerTurnResult{Success: false}, err
				}
				MonsterTurnSystem(world, eventLogList)
				// TODO: emit events to client
				fmt.Println("TODO: emit activated abilities and spell log to client")

			case "wand":
				wandnum, err := strconv.Atoi(turn.Msg.WandNum)
				if err != nil {
					return msg.PlayerTurnResult{}, fmt.Errorf("error converting string to int: %w", err)
				}
				castID, potentialAbilities, err := player_turn_wand(world, direction, wandnum)
				if err != nil {
					return msg.PlayerTurnResult{Success: false}, err
				}
				fmt.Println("castID: ", castID)
				fmt.Println("potentialAbilities: ", potentialAbilities)

				fmt.Println("gameidstr:", turn.Msg.GameIDStr)

				gameID, err := strconv.Atoi(turn.Msg.GameIDStr)
				if err != nil {
					return msg.PlayerTurnResult{}, fmt.Errorf("error converting string to int: %w", err)
				}
				revealRequest := client.RevealRequest{
					PersonaTag:         turn.Tx.PersonaTag,
					GameID:             types.EntityID(gameID),
					CastID:             castID,
					WandNum:            wandnum,
					PotentialAbilities: *potentialAbilities,
				}
				revealRequest.PotentialAbilities = [2]bool{true, true}
				revealRequestCh <- revealRequest
				fmt.Println("PlayerTurnSystem *potentialAbilities", revealRequest.PotentialAbilities)

			case "move":
				err = playerTurnMove(world, direction)
				if err != nil {
					return msg.PlayerTurnResult{}, fmt.Errorf("PlayerTurnSystem err: %w", err)
				}
				MonsterTurnSystem(world, eventLogList)
				fmt.Println("TODO: emit activated abilities and spell log to client")
			default:
				return msg.PlayerTurnResult{}, fmt.Errorf("PlayerTurnSystem err: Invalid action")
			}

			err = world.EmitEvent(map[string]any{
				"event":     "player_turn",
				"action":    turn.Msg.Action,
				"direction": direction,
			})
			if err != nil {
				return msg.PlayerTurnResult{}, err
			}

			result := msg.PlayerTurnResult{Success: true}
			return result, nil

		})
}

func player_turn_attack(world cardinal.WorldContext, direction comp.Direction) error {
	playerPos, err := cardinal.GetComponent[comp.Position](world, 0)
	if err != nil {
		return err
	}
	attackPos, err := playerPos.GetUpdateFromDirection(direction)
	if err != nil {
		return err
	}

	found, id, err := attackPos.GetEntityIDByPosition(world)
	if err != nil {
		return err
	}
	if found {
		colType, err := cardinal.GetComponent[comp.Collidable](world, id)
		if err != nil {
			return err
		}
		switch colType.Type {
		case comp.MonsterCollide:
			return comp.DecrementHealth(world, id)
		default:
			return fmt.Errorf("attempting to attack %s", colType.ToString())
		}
	} else {
		return errors.New("attempting to attack empty stace")
	}
}

func player_turn_wand(world cardinal.WorldContext, direction comp.Direction, wandnum int) (castID types.EntityID, potentialAbilities *[client.TotalAbilities]bool, err error) {
	playerPos, err := cardinal.GetComponent[comp.Position](world, 0)
	if err != nil {
		return 0, nil, err
	}
	spellPos, err := playerPos.GetUpdateFromDirection(direction)
	if err != nil {
		return 0, nil, err
	}
	wandID, _, available, err := getWandByNumber(world, wandnum)
	if err != nil {
		return 0, nil, err
	}

	// handle wand availability
	if !available.IsAvailable {
		return 0, nil, fmt.Errorf("wand %d already expired", wandnum)
	}
	// set the wand to not ready (do early as it may potentially be refreshed by abilities)
	cardinal.SetComponent[comp.Available](world, wandID, &comp.Available{IsAvailable: false})

	// set all abilities to true since we don't know which ones will be activated
	allAbilities := &[client.TotalAbilities]bool{}
	for i := range allAbilities {
		allAbilities[i] = true
	}
	spell := &comp.Spell{
		WandNumber: wandnum,
		Expired:    false,
		Abilities:  allAbilities,
		Direction:  direction,
	}

	// simulate a cast to determine potential ability activations
	updateChainState := false
	dummy := &[]comp.GameEventLog{} // dummy event log, not used for anything but to satisfy the function signature
	err = resolveAbilities(world, spell, spellPos, spell.Abilities, updateChainState, dummy)
	if err != nil {
		return 0, nil, err
	}

	// create a new entity for the cast to later be resolved
	castID, err = cardinal.Create(
		world,
		comp.AwaitingReveal{IsAvailable: true},
		spell,
		spellPos,
	)
	if err != nil {
		return 0, nil, err
	}

	return castID, spell.Abilities, nil

}

func playerTurnMove(world cardinal.WorldContext, direction comp.Direction) error {
	playerID, err := queryPlayerID(world)
	if err != nil {
		return err
	}
	currPos, err := cardinal.GetComponent[comp.Position](world, playerID)
	if err != nil {
		return err
	}
	updatePos, err := currPos.GetUpdateFromDirection(direction)
	if err != nil {
		return err
	}

	found, id, err := updatePos.GetEntityIDByPosition(world)
	if err != nil {
		return err
	}
	if found {
		colType, err := cardinal.GetComponent[comp.Collidable](world, id)
		if err != nil {
			return err
		}
		if colType.Type != comp.ItemCollide {
			return fmt.Errorf("attempting to move onto an object of type %s", colType.ToString())
		}
	}

	cardinal.SetComponent[comp.Position](world, playerID, updatePos)

	return nil
}
