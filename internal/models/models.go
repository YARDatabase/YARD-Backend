package models

import "time"

type HealthResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type HypixelAPIResponse struct {
	Success     bool   `json:"success"`
	LastUpdated int64  `json:"lastUpdated"`
	Items       []Item `json:"items"`
}

type BazaarOrder struct {
	Amount      int64   `json:"amount"`
	PricePerUnit float64 `json:"pricePerUnit"`
	Orders      int     `json:"orders"`
}

type ReforgeStats struct {
	Health            *float64 `json:"health,omitempty"`
	Defense            *float64 `json:"defense,omitempty"`
	Strength          *float64 `json:"strength,omitempty"`
	Intelligence      *float64 `json:"intelligence,omitempty"`
	CritChance        *float64 `json:"crit_chance,omitempty"`
	CritDamage        *float64 `json:"crit_damage,omitempty"`
	AttackSpeed       *float64 `json:"attack_speed,omitempty"`
	BonusAttackSpeed  *float64 `json:"bonus_attack_speed,omitempty"`
	Speed             *float64 `json:"speed,omitempty"`
	MiningSpeed       *float64 `json:"mining_speed,omitempty"`
	MiningFortune     *float64 `json:"mining_fortune,omitempty"`
	FarmingFortune    *float64 `json:"farming_fortune,omitempty"`
	Damage            *float64 `json:"damage,omitempty"`
	SeaCreatureChance *float64 `json:"sea_creature_chance,omitempty"`
	MagicFind         *float64 `json:"magic_find,omitempty"`
	PetLuck           *float64 `json:"pet_luck,omitempty"`
	TrueDefense       *float64 `json:"true_defense,omitempty"`
	Ferocity          *float64 `json:"ferocity,omitempty"`
	AbilityDamage     *float64 `json:"ability_damage,omitempty"`
}

type ReforgeEffect struct {
	ReforgeName      string                  `json:"reforge_name,omitempty"`
	ItemTypes        string                  `json:"item_types,omitempty"`
	RequiredRarities []string                `json:"required_rarities,omitempty"`
	ReforgeStats     map[string]ReforgeStats `json:"reforge_stats,omitempty"`
	ReforgeAbility   interface{}            `json:"reforge_ability,omitempty"`
	ReforgeCosts     map[string]int          `json:"reforge_costs,omitempty"`
	Description      []string                `json:"description,omitempty"`
	Obtaining        string                  `json:"obtaining,omitempty"`
	MiningLevelReq   string                  `json:"mining_level_req,omitempty"`
}

type Item struct {
	Name            string                 `json:"name"`
	Category        string                 `json:"category"`
	Tier            string                 `json:"tier"`
	ID              string                 `json:"id"`
	NPCSellPrice    interface{}            `json:"npc_sell_price,omitempty"`
	Stats           map[string]interface{} `json:"stats,omitempty"`
	Skin            map[string]interface{} `json:"skin,omitempty"`
	Glowing         bool                   `json:"glowing,omitempty"`
	Soulbound       string                 `json:"soulbound,omitempty"`
	Requirements    []interface{}          `json:"requirements,omitempty"`
	CanAuction      bool                   `json:"can_auction,omitempty"`
	ItemSpecific    map[string]interface{} `json:"item_specific,omitempty"`
	AuctionPrice    *int64                 `json:"auction_price,omitempty"`
	BazaarBuyPrice  *float64               `json:"bazaar_buy_price,omitempty"`
	BazaarSellPrice *float64               `json:"bazaar_sell_price,omitempty"`
	BazaarBuyOrders []BazaarOrder          `json:"bazaar_buy_orders,omitempty"`
	BazaarSellOrders []BazaarOrder         `json:"bazaar_sell_orders,omitempty"`
	ReforgeEffect   *ReforgeEffect         `json:"reforge_effect,omitempty"`
}

type ReforgeStonesResponse struct {
	Success      bool      `json:"success"`
	Count        int       `json:"count"`
	LastUpdated  time.Time `json:"lastUpdated"`
	ReforgeStones []Item   `json:"reforgeStones"`
}

// reforge represents a complete reforge from the neu reforges.json file
type Reforge struct {
	ReforgeName      string                  `json:"reforge_name"`
	ItemTypes        string                  `json:"item_types"`
	RequiredRarities []string                `json:"required_rarities"`
	ReforgeStats     map[string]ReforgeStats `json:"reforge_stats"`
	ReforgeAbility   interface{}             `json:"reforge_ability,omitempty"`
	ReforgeCosts     map[string]int          `json:"reforge_costs,omitempty"`
	Source           string                  `json:"source"`
	StoneID          string                  `json:"stone_id,omitempty"`
	StoneName        string                  `json:"stone_name,omitempty"`
	StoneTier        string                  `json:"stone_tier,omitempty"`
	StonePrice       *int64                  `json:"stone_price,omitempty"`
}

// reforgesresponse is the api response containing all reforges
type ReforgesResponse struct {
	Success     bool      `json:"success"`
	Count       int       `json:"count"`
	LastUpdated time.Time `json:"lastUpdated"`
	Reforges    []Reforge `json:"reforges"`
}

