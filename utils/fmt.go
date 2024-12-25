package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatNumber(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}

	f := float32(n) / 1000

	if f < 1000 {
		return fmt.Sprintf("%.1fK", f)
	}

	f = f / 1000

	if f < 1000 {
		return fmt.Sprintf("%.1fM", f)
	}

	f = f / 1000

	if f < 1000 {
		return fmt.Sprintf("%.1fB", f)
	}

	f = f / 1000

	if f < 1000 {
		return fmt.Sprintf("%.1fT", f)
	}

	f = f / 1000

	return fmt.Sprintf("%.1fQ", f)
}

func FormatDamageLabel(health, armor, power int) string {
	strs := make([]string, 0, 3)

	if health != 0 {
		strs = append(strs, fmt.Sprintf("%v HP", FormatNumber(health)))
	}

	if armor != 0 {
		strs = append(strs, fmt.Sprintf("%v AP", FormatNumber(armor)))
	}

	if power != 0 {
		strs = append(strs, fmt.Sprintf("%v P", FormatNumber(power)))
	}

	if health == 0 && armor == 0 && power == 0 {
		strs = append(strs, "None!")
	}

	return strings.Join(strs, ", ")
}
