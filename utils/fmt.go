package utils

import (
	"fmt"
	"strings"
)

func FormatDamageLabel(health, armor, power int) string {
	strs := make([]string, 0, 3)

	if health != 0 {
		strs = append(strs, fmt.Sprintf("%v HP", health))
	}

	if armor != 0 {
		strs = append(strs, fmt.Sprintf("%v AP", armor))
	}

	if power != 0 {
		strs = append(strs, fmt.Sprintf("%v P", power))
	}

	if health == 0 && armor == 0 && power == 0 {
		strs = append(strs, "None!")
	}

	return strings.Join(strs, ", ")
}
