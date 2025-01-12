package abstract

type Settings struct {
	Theme                   PGCTThemeSelection
	TickIntervalSeconds     float64
	SecondsUntilDPSReset    int
	RemoveLevelsFromSkills  bool
	EntitiesThatCountAsPets []string
}

func NewSettings() *Settings {
	return &Settings{
		TickIntervalSeconds:    2,
		SecondsUntilDPSReset:   15,
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
