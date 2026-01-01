package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"yard-backend/internal/config"
	"yard-backend/internal/models"
	"yard-backend/internal/utils"
)

// loads reforge stone definitions from the notenoughupdates repository json file
func LoadNEUReforgeStones() error {
	config.NEUReforgeStonesMutex.Lock()
	defer config.NEUReforgeStonesMutex.Unlock()
	
	path := fmt.Sprintf("%s/constants/reforgestones.json", config.NEURepoPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read reforgestones.json: %w", err)
	}
	
	var reforgestones map[string]interface{}
	if err := json.Unmarshal(data, &reforgestones); err != nil {
		return fmt.Errorf("failed to parse reforgestones.json: %w", err)
	}
	
	config.NEUReforgeStones = reforgestones
	log.Printf("Loaded %d reforge stone definitions from NEU", len(config.NEUReforgeStones))
	return nil
}

// gets item lore data from the notenoughupdates repository for a specific item id
func GetNEUItemData(itemID string) ([]string, error) {
	path := fmt.Sprintf("%s/items/%s.json", config.NEURepoPath, itemID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var itemData struct {
		Lore []string `json:"lore"`
	}
	if err := json.Unmarshal(data, &itemData); err != nil {
		return nil, err
	}
	
	return itemData.Lore, nil
}

// gets the reforge effect data for a stone including stats costs abilities and descriptions
func GetReforgeEffectForStone(itemID string) *models.ReforgeEffect {
	config.NEUReforgeStonesMutex.RLock()
	defer config.NEUReforgeStonesMutex.RUnlock()
	
	stoneData, exists := config.NEUReforgeStones[itemID]
	if !exists {
		return nil
	}
	
	stoneMap, ok := stoneData.(map[string]interface{})
	if !ok {
		return nil
	}
	
	effect := &models.ReforgeEffect{}
	
	if reforgeName, ok := stoneMap["reforgeName"].(string); ok {
		effect.ReforgeName = reforgeName
	}
	if itemTypes, ok := stoneMap["itemTypes"].(string); ok {
		effect.ItemTypes = itemTypes
	}
	if rarities, ok := stoneMap["requiredRarities"].([]interface{}); ok {
		effect.RequiredRarities = make([]string, len(rarities))
		for i, r := range rarities {
			if rStr, ok := r.(string); ok {
				effect.RequiredRarities[i] = rStr
			}
		}
	}
	if costs, ok := stoneMap["reforgeCosts"].(map[string]interface{}); ok {
		effect.ReforgeCosts = make(map[string]int)
		for k, v := range costs {
			if costFloat, ok := v.(float64); ok {
				effect.ReforgeCosts[k] = int(costFloat)
			}
		}
	}
	if ability, ok := stoneMap["reforgeAbility"]; ok {
		effect.ReforgeAbility = ability
	}
	if stats, ok := stoneMap["reforgeStats"].(map[string]interface{}); ok {
		effect.ReforgeStats = make(map[string]models.ReforgeStats)
		for rarity, statData := range stats {
			if statMap, ok := statData.(map[string]interface{}); ok {
				reforgeStat := models.ReforgeStats{}
				if v, ok := statMap["health"].(float64); ok {
					reforgeStat.Health = &v
				}
				if v, ok := statMap["defense"].(float64); ok {
					reforgeStat.Defense = &v
				}
				if v, ok := statMap["strength"].(float64); ok {
					reforgeStat.Strength = &v
				}
				if v, ok := statMap["intelligence"].(float64); ok {
					reforgeStat.Intelligence = &v
				}
				if v, ok := statMap["crit_chance"].(float64); ok {
					reforgeStat.CritChance = &v
				}
				if v, ok := statMap["crit_damage"].(float64); ok {
					reforgeStat.CritDamage = &v
				}
				if v, ok := statMap["attack_speed"].(float64); ok {
					reforgeStat.AttackSpeed = &v
				}
				if v, ok := statMap["bonus_attack_speed"].(float64); ok {
					reforgeStat.BonusAttackSpeed = &v
				}
				if v, ok := statMap["speed"].(float64); ok {
					reforgeStat.Speed = &v
				}
				if v, ok := statMap["mining_speed"].(float64); ok {
					reforgeStat.MiningSpeed = &v
				}
				if v, ok := statMap["mining_fortune"].(float64); ok {
					reforgeStat.MiningFortune = &v
				}
				if v, ok := statMap["farming_fortune"].(float64); ok {
					reforgeStat.FarmingFortune = &v
				}
				if v, ok := statMap["damage"].(float64); ok {
					reforgeStat.Damage = &v
				}
				if v, ok := statMap["sea_creature_chance"].(float64); ok {
					reforgeStat.SeaCreatureChance = &v
				}
				if v, ok := statMap["magic_find"].(float64); ok {
					reforgeStat.MagicFind = &v
				}
				if v, ok := statMap["pet_luck"].(float64); ok {
					reforgeStat.PetLuck = &v
				}
				if v, ok := statMap["true_defense"].(float64); ok {
					reforgeStat.TrueDefense = &v
				}
				if v, ok := statMap["ferocity"].(float64); ok {
					reforgeStat.Ferocity = &v
				}
				if v, ok := statMap["ability_damage"].(float64); ok {
					reforgeStat.AbilityDamage = &v
				}
				effect.ReforgeStats[rarity] = reforgeStat
			}
		}
	}
	
	lore, err := GetNEUItemData(itemID)
	if err == nil && len(lore) > 0 {
		effect.Description = lore
		effect.Obtaining = utils.ExtractObtainingFromLore(lore)
		effect.MiningLevelReq = utils.ExtractMiningLevelFromLore(lore)
	}
	
	return effect
}

