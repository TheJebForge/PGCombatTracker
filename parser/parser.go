package parser

import (
	"PGCombatTracker/abstract"
	"strconv"
	"strings"
	"time"
)

const timeFormat = "06-01-02 15:04:05"

func ParseLine(line string) (*abstract.ChatEvent, error) {
	timeString, rest, found := strings.Cut(line, "\t")

	if !found {
		return nil, nil
	}

	timeValue, err := time.Parse(timeFormat, timeString)

	if err != nil {
		return nil, err
	}

	// We got Time, look for combat lines
	if strings.HasPrefix(rest, "[Combat]") {
		_, rest, found = strings.Cut(rest, " ")

		if !found {
			return nil, nil
		}

		subject, rest, found := strings.Cut(rest, ": ")

		if !found {
			return nil, nil
		}

		skill, rest, found := strings.Cut(rest, " on ")

		if found {
			skillUse, err := parseSkillUse(subject, skill, rest)

			if err != nil {
				return nil, err
			}

			return &abstract.ChatEvent{
				Time:     timeValue,
				Contents: skillUse,
			}, nil
		}

		left, rest, found := strings.Cut(skill, "Recovered: ")

		if found {
			recovered, err := parseRecovered(subject, rest)

			if err != nil {
				return nil, err
			}

			return &abstract.ChatEvent{
				Time:     timeValue,
				Contents: recovered,
			}, nil
		}

		_, rest, found = strings.Cut(left, "Suffered indirect dmg: ")

		if found {
			indirect, err := parseIndirectDamage(subject, rest)

			if err != nil {
				return nil, err
			}

			return &abstract.ChatEvent{
				Time:     timeValue,
				Contents: indirect,
			}, nil
		}
	} else if strings.HasPrefix(rest, "[Status] You earned ") {
		return parseXPGain(timeValue, rest)
	} else if strings.HasPrefix(rest, "[Status] You searched the corpse and found ") {
		return parseFoundCoins(timeValue, rest)
	} else if strings.HasPrefix(rest, "[Status] You receive ") {
		return parseReceivedCoins(timeValue, rest)
	} else if strings.HasPrefix(rest, "***") {
		return parseLogin(timeValue, rest)
	} else if strings.HasPrefix(rest, "[Error]") {
		_, rest, _ := strings.Cut(rest, "[Error] ")
		return &abstract.ChatEvent{
			Time: timeValue,
			Contents: &abstract.ErrorLine{
				Message: rest,
			},
		}, nil
	}

	return nil, nil
}

func parseLogin(timeValue time.Time, rest string) (*abstract.ChatEvent, error) {
	_, rest, found := strings.Cut(rest, "* Logged In As ")

	if !found {
		return nil, nil
	}

	name, _, found := strings.Cut(rest, ".")

	if !found {
		return nil, nil
	}

	return &abstract.ChatEvent{
		Time: timeValue,
		Contents: &abstract.Login{
			Name: name,
		},
	}, nil
}

func parseSkillUse(subject, skill, rest string) (*abstract.SkillUse, error) {
	var (
		victim string
		found  bool
		crit   bool
		evaded bool
	)

	if strings.Contains(rest, "(CRIT!)") {
		victim, rest, _ = strings.Cut(rest, " (CRIT!)")
		crit = true
	} else if strings.Contains(rest, "(EVADED!)") {
		victim, rest, _ = strings.Cut(rest, " (EVADED!)")
		evaded = true
	} else {
		victim, rest, found = strings.Cut(rest, "!")

		if !found {
			return nil, nil
		}
	}

	fatality := strings.Contains(rest, "(FATALITY!)")

	_, rest, found = strings.Cut(rest, "Dmg: ")

	if !found {
		return &abstract.SkillUse{
			Subject:  subject,
			Skill:    skill,
			Victim:   victim,
			Crit:     crit,
			Evaded:   evaded,
			Fatality: fatality,
		}, nil
	}

	if strings.Contains(rest, "none") {
		return &abstract.SkillUse{
			Subject:  subject,
			Skill:    skill,
			Victim:   victim,
			Crit:     crit,
			Evaded:   evaded,
			Fatality: fatality,
			Damage:   &abstract.Vitals{},
		}, nil
	}

	health, rest, found := strings.Cut(rest, " health")
	if !found {
		rest = health
	}

	// get rid of comma
	left, rest, found := strings.Cut(rest, ", ")
	if !found {
		rest = left
	}

	armor, rest, found := strings.Cut(rest, " armor")
	if !found {
		rest = armor
	}

	left, rest, found = strings.Cut(rest, ", ")
	if !found {
		rest = left
	}

	power, rest, _ := strings.Cut(rest, " power")

	healthValue, _ := strconv.Atoi(health)
	armorValue, _ := strconv.Atoi(armor)
	powerValue, _ := strconv.Atoi(power)

	return &abstract.SkillUse{
		Subject: subject,
		Skill:   skill,
		Victim:  victim,
		Crit:    crit,
		Evaded:  evaded,
		Damage: &abstract.Vitals{
			Health: healthValue,
			Armor:  armorValue,
			Power:  powerValue,
		},
		Fatality: fatality,
	}, nil
}

func parseRecovered(subject, rest string) (*abstract.Recovered, error) {
	health, rest, found := strings.Cut(rest, " health")
	if !found {
		rest = health
	}

	// get rid of comma
	left, rest, found := strings.Cut(rest, ", ")
	if !found {
		rest = left
	}

	armor, rest, found := strings.Cut(rest, " armor")
	if !found {
		rest = armor
	}

	left, rest, found = strings.Cut(rest, ", ")
	if !found {
		rest = left
	}

	power, rest, _ := strings.Cut(rest, " power")

	healthValue, _ := strconv.Atoi(health)
	armorValue, _ := strconv.Atoi(armor)
	powerValue, _ := strconv.Atoi(power)

	return &abstract.Recovered{
		Subject: subject,
		Healed: abstract.Vitals{
			Health: healthValue,
			Armor:  armorValue,
			Power:  powerValue,
		},
	}, nil
}

func parseIndirectDamage(subject, rest string) (*abstract.IndirectDamage, error) {
	health, rest, found := strings.Cut(rest, " health")
	if !found {
		rest = health
	}

	// get rid of comma
	left, rest, found := strings.Cut(rest, ", ")
	if !found {
		rest = left
	}

	armor, rest, found := strings.Cut(rest, " armor")
	if !found {
		rest = armor
	}

	left, rest, found = strings.Cut(rest, ", ")
	if !found {
		rest = left
	}

	power, rest, _ := strings.Cut(rest, " power")

	healthValue, _ := strconv.Atoi(health)
	armorValue, _ := strconv.Atoi(armor)
	powerValue, _ := strconv.Atoi(power)

	return &abstract.IndirectDamage{
		Subject: subject,
		Damage: abstract.Vitals{
			Health: healthValue,
			Armor:  armorValue,
			Power:  powerValue,
		},
	}, nil
}

func parseXPGain(time time.Time, line string) (*abstract.ChatEvent, error) {
	_, rest, found := strings.Cut(line, "You earned ")

	if !found {
		return nil, nil
	}

	xp, rest, found := strings.Cut(rest, " XP in ")

	if !found {
		xp, rest, found = strings.Cut(xp, " XP and reached level ")
		xpValue, _ := strconv.Atoi(xp)

		if !found {
			return nil, nil
		}

		level, rest, found := strings.Cut(rest, " in ")
		levelValue, _ := strconv.Atoi(level)

		if !found {
			return nil, nil
		}

		skill, _, _ := strings.Cut(rest, "!")

		return &abstract.ChatEvent{
			Time: time,
			Contents: &abstract.XPGainedLeveledUp{
				XP:    xpValue,
				Skill: skill,
				Level: levelValue,
			},
		}, nil
	}

	xpValue, _ := strconv.Atoi(xp)

	skill, _, _ := strings.Cut(rest, ".")

	return &abstract.ChatEvent{
		Time: time,
		Contents: &abstract.XPGained{
			XP:    xpValue,
			Skill: skill,
		},
	}, nil
}

func parseFoundCoins(time time.Time, line string) (*abstract.ChatEvent, error) {
	_, rest, found := strings.Cut(line, "You searched the corpse and found ")

	if !found {
		return nil, nil
	}

	coins, _, found := strings.Cut(rest, " coins")

	if !found {
		return nil, nil
	}

	coinsValue, _ := strconv.Atoi(coins)

	return &abstract.ChatEvent{
		Time: time,
		Contents: &abstract.FoundCoins{
			Coins: coinsValue,
		},
	}, nil
}

func parseReceivedCoins(time time.Time, line string) (*abstract.ChatEvent, error) {
	_, rest, found := strings.Cut(line, "You receive ")

	if !found {
		return nil, nil
	}

	coins, _, found := strings.Cut(rest, " coins")

	if !found {
		return nil, nil
	}

	coinsValue, _ := strconv.Atoi(coins)

	return &abstract.ChatEvent{
		Time: time,
		Contents: &abstract.ReceivedCoins{
			Coins: coinsValue,
		},
	}, nil
}
