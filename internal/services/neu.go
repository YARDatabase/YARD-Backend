package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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
			if v, ok := statMap["foraging_fortune"].(float64); ok {
				reforgeStat.ForagingFortune = &v
			}
			if v, ok := statMap["foraging_wisdom"].(float64); ok {
				reforgeStat.ForagingWisdom = &v
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

// loads all reforge definitions from the notenoughupdates repository reforges.json file
func LoadNEUReforges() error {
	config.NEUReforgesMutex.Lock()
	defer config.NEUReforgesMutex.Unlock()
	
	path := fmt.Sprintf("%s/constants/reforges.json", config.NEURepoPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read reforges.json: %w", err)
	}
	
	var reforges map[string]interface{}
	if err := json.Unmarshal(data, &reforges); err != nil {
		return fmt.Errorf("failed to parse reforges.json: %w", err)
	}
	
	config.NEUReforges = reforges
	log.Printf("Loaded %d reforge definitions from NEU reforges.json", len(config.NEUReforges))
	return nil
}

// gets all reforges merging data from reforges.json and reforgestones.json
func GetAllReforges() []models.Reforge {
	config.NEUReforgesMutex.RLock()
	defer config.NEUReforgesMutex.RUnlock()
	
	config.NEUReforgeStonesMutex.RLock()
	defer config.NEUReforgeStonesMutex.RUnlock()
	
	reforgeMap := make(map[string]*models.Reforge)
	
	// first load all reforges from reforges.json (blacksmith reforges)
	for reforgeName, reforgeData := range config.NEUReforges {
		reforgeDataMap, ok := reforgeData.(map[string]interface{})
		if !ok {
			continue
		}
		
		reforge := parseReforgeData(reforgeName, reforgeDataMap, "Blacksmith")
		if reforge != nil {
			reforgeMap[reforgeName] = reforge
		}
	}
	
	// then overlay/add reforges from reforgestones.json (stone reforges have priority)
	for stoneID, stoneData := range config.NEUReforgeStones {
		stoneDataMap, ok := stoneData.(map[string]interface{})
		if !ok {
			continue
		}
		
		reforgeName, _ := stoneDataMap["reforgeName"].(string)
		if reforgeName == "" {
			continue
		}
		
		reforge := parseReforgeData(reforgeName, stoneDataMap, "Reforge Stone")
		if reforge != nil {
			reforge.StoneID = stoneID
			
			// fetch stone price and details from redis
			if config.RDB != nil {
				key := fmt.Sprintf("reforge_stone:%s", stoneID)
				stoneJSON, err := config.RDB.Get(config.Ctx, key).Result()
				if err == nil && stoneJSON != "" {
					var stone models.Item
					if err := json.Unmarshal([]byte(stoneJSON), &stone); err == nil {
						reforge.StoneName = stone.Name
						reforge.StoneTier = stone.Tier
						
						// get best available price (auction > bazaar buy > bazaar sell)
						if stone.AuctionPrice != nil {
							price := int64(*stone.AuctionPrice)
							reforge.StonePrice = &price
						} else if stone.BazaarBuyPrice != nil {
							price := int64(*stone.BazaarBuyPrice)
							reforge.StonePrice = &price
						} else if stone.BazaarSellPrice != nil {
							price := int64(*stone.BazaarSellPrice)
							reforge.StonePrice = &price
						}
					}
				}
			}
			
			reforgeMap[reforgeName] = reforge
		}
	}
	
	// convert map to slice
	reforges := make([]models.Reforge, 0, len(reforgeMap))
	for _, reforge := range reforgeMap {
		reforges = append(reforges, *reforge)
	}
	
	// apply manual data corrections for known NEU data issues
	applyDataCorrections(reforges)
	
	return reforges
}

// applies manual corrections to fix known data issues in NEU repo
func applyDataCorrections(reforges []models.Reforge) {
	for i := range reforges {
		switch reforges[i].ReforgeName {
		case "Ancient":
			// fix: COMMON tier has crit_damage instead of crit_chance
			// the +3 value should be crit_chance, not crit_damage
			if stats, ok := reforges[i].ReforgeStats["COMMON"]; ok {
				if stats.CritDamage != nil && *stats.CritDamage == 3 {
					critChance := float64(3)
					stats.CritChance = &critChance
					stats.CritDamage = nil
					reforges[i].ReforgeStats["COMMON"] = stats
				}
			}
		}
	}
}

// parses reforge data from a map into a reforge struct
func parseReforgeData(reforgeName string, data map[string]interface{}, source string) *models.Reforge {
	reforge := &models.Reforge{
		ReforgeName: reforgeName,
		Source:      source,
	}
	
	// parse item types - can be string or object with internalName array
	if itemTypes, ok := data["itemTypes"].(string); ok {
		reforge.ItemTypes = itemTypes
	} else if itemTypesObj, ok := data["itemTypes"].(map[string]interface{}); ok {
		// special item restrictions like specific items
		if internalNames, ok := itemTypesObj["internalName"].([]interface{}); ok {
			names := make([]string, 0, len(internalNames))
			for _, name := range internalNames {
				if nameStr, ok := name.(string); ok {
					names = append(names, nameStr)
				}
			}
			reforge.ItemTypes = "SPECIFIC:" + strings.Join(names, ",")
		} else if itemIds, ok := itemTypesObj["itemId"].([]interface{}); ok {
			ids := make([]string, 0, len(itemIds))
			for _, id := range itemIds {
				if idStr, ok := id.(string); ok {
					ids = append(ids, idStr)
				}
			}
			reforge.ItemTypes = "SPECIFIC:" + strings.Join(ids, ",")
		}
	}
	
	// parse required rarities
	if rarities, ok := data["requiredRarities"].([]interface{}); ok {
		reforge.RequiredRarities = make([]string, 0, len(rarities))
		for _, r := range rarities {
			if rStr, ok := r.(string); ok {
				reforge.RequiredRarities = append(reforge.RequiredRarities, rStr)
			}
		}
	}
	
	// parse reforge costs
	if costs, ok := data["reforgeCosts"].(map[string]interface{}); ok {
		reforge.ReforgeCosts = make(map[string]int)
		for k, v := range costs {
			if costFloat, ok := v.(float64); ok {
				reforge.ReforgeCosts[k] = int(costFloat)
			}
		}
	}
	
	// parse reforge ability
	if ability, ok := data["reforgeAbility"]; ok {
		reforge.ReforgeAbility = ability
	}
	
	// parse reforge stats
	if stats, ok := data["reforgeStats"].(map[string]interface{}); ok {
		reforge.ReforgeStats = make(map[string]models.ReforgeStats)
		for rarity, statData := range stats {
			if statMap, ok := statData.(map[string]interface{}); ok {
				reforgeStat := parseReforgeStats(statMap)
				reforge.ReforgeStats[rarity] = reforgeStat
			}
		}
	}
	
	return reforge
}

// parses stat values from a map into a reforgestats struct
func parseReforgeStats(statMap map[string]interface{}) models.ReforgeStats {
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
			if v, ok := statMap["foraging_fortune"].(float64); ok {
				reforgeStat.ForagingFortune = &v
			}
			if v, ok := statMap["foraging_wisdom"].(float64); ok {
				reforgeStat.ForagingWisdom = &v
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
	
	return reforgeStat
}

