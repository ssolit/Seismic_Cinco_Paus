package component

import (
	"pkg.world.dev/world-engine/cardinal"
	"pkg.world.dev/world-engine/cardinal/types"
)

// polymorph
const Ability10ID = 10

type Ability10 struct{}

var _ Ability = &Ability10{}

func (Ability10) GetAbilityID() int {
	return Ability10ID
}
func (Ability10) GetAbilityName() string {
	return "polymorph"
}

// polymorphs the monster
func (Ability10) Resolve(
	world cardinal.WorldContext,
	gameID types.EntityID,
	spellPosition *Position,
	_ Direction,
	executeUpdates bool,
	eventLogList *[]GameEventLog,
) (reveal bool, err error) {
	// Lookup if entity exists
	found, id, err := spellPosition.GetEntityIDByPosition(world, gameID)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	// check if its a monster
	colType, err := cardinal.GetComponent[Collidable](world, id)
	if err != nil {
		return false, err
	}
	if colType.Type != MonsterCollide {
		return false, nil
	}

	if executeUpdates {
		newMonsterType, err := polymorphMonster(world, gameID, id)
		if err != nil {
			// log.Println("polymorphMonster() err: %w", err)
			return false, err
		}
		monsterTypeToPolymorphEventOffset := 15
		polymorphEvent := GameEvent(int(newMonsterType) + monsterTypeToPolymorphEventOffset)
		gameEvent := GameEventLog{X: spellPosition.X, Y: spellPosition.Y, Event: polymorphEvent}
		*eventLogList = append(*eventLogList, gameEvent)
	}

	// hit a monster, so ability should reveal
	return true, nil
}

func polymorphMonster(world cardinal.WorldContext, gameID types.EntityID, monID types.EntityID) (newMonsterType MonsterType, err error) {
	// get monster type
	monster, err := cardinal.GetComponent[Monster](world, monID)
	if err != nil {
		return 0, err
	}
	newMonsterType = (monster.Type + 1) % NumMonsterTypes

	// get monster pos and health
	monsterPos, err := cardinal.GetComponent[Position](world, monID)
	if err != nil {
		return 0, err
	}
	oldMonsterHealth, err := cardinal.GetComponent[Health](world, monID)
	if err != nil {
		return 0, err
	}
	oldMonsterDamage := oldMonsterHealth.MaxHealth - oldMonsterHealth.CurrHealth

	err = cardinal.Remove(world, monID)
	if err != nil {
		return 0, err
	}

	// create new monster
	_, err = cardinal.Create(world,
		Monster{
			Type: newMonsterType,
		},
		Collidable{Type: MonsterCollide},
		Health{
			MaxHealth:  int(newMonsterType) + 1,
			CurrHealth: int(newMonsterType) + 1 - oldMonsterDamage,
		},
		Position{
			X: monsterPos.X,
			Y: monsterPos.Y,
		},
		GameObj{GameID: gameID},
	)
	if err != nil {
		return 0, err
	}
	// log.Println("oldMonID: ", monID)
	// log.Println("newMonID: ", newMonID)

	return newMonsterType, nil

}
