package utils

import "strings"

// extracts the obtaining method text from item lore by searching for keywords
func ExtractObtainingFromLore(lore []string) string {
	for i, line := range lore {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "obtained") || 
		   strings.Contains(lineLower, "drop") || 
		   strings.Contains(lineLower, "found") ||
		   strings.Contains(lineLower, "purchase") ||
		   strings.Contains(lineLower, "craft") {
			if i+1 < len(lore) && !strings.HasPrefix(lore[i+1], "§") {
				return strings.TrimSpace(strings.TrimPrefix(lore[i+1], "§7"))
			}
			return strings.TrimSpace(strings.ReplaceAll(line, "§7", ""))
		}
	}
	return ""
}

// extracts mining level requirement text from item lore
func ExtractMiningLevelFromLore(lore []string) string {
	for _, line := range lore {
		if strings.Contains(line, "Mining Skill Level") || strings.Contains(line, "Mining Level") {
			line = strings.ReplaceAll(line, "§7", "")
			line = strings.ReplaceAll(line, "§a", "")
			line = strings.ReplaceAll(line, "!", "")
			line = strings.TrimSpace(line)
			return line
		}
	}
	return ""
}

