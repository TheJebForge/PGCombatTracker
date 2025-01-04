package abstract

type Settings struct {
	RemoveLevelsFromSkills  bool
	EntitiesThatCountAsPets []string
}

func NewSettings() *Settings {
	return &Settings{
		RemoveLevelsFromSkills: true,
		EntitiesThatCountAsPets: []string{
			"Summoned Golem Minion",
			"Armored Golem Minion",
			"Addled Figment",
			"Explosion Trap",
			"Electricity Sigil",
			"Acid Sigil",
			"Summoned Deer",
			"Cold Sphere",
			"Coldfire Wall",
			"Smarmfire Wall",
			"Healing Flame Wall",
			"Fire Wall",
			"Tornado",
			"Sandstorm",
			"Doomstorm",
		},
	}
}
