package abstract

import (
	"PGCombatTracker/utils"
	"fmt"
	"time"
)

type ChatEvent struct {
	Time     time.Time
	Contents ChatContent
}

// ChatContent Just a marker, so only chat stuff can go into chat event
type ChatContent interface {
	ImplementsChatContent()
}

type Login struct {
	Name string
}

type SkillUse struct {
	Subject  string
	Skill    string
	Victim   string
	Damage   *Vitals
	Evaded   bool
	Crit     bool
	Fatality bool
}

type Vitals struct {
	Health int
	Armor  int
	Power  int
}

type Recovered struct {
	Subject string
	Healed  Vitals
}

type IndirectDamage struct {
	Subject string
	Damage  Vitals
}

type XPGained struct {
	XP    int
	Skill string
}

type XPGainedLeveledUp struct {
	XP    int
	Skill string
	Level int
}

type FoundCoins struct {
	Coins int
}

type ReceivedCoins struct {
	Coins int
}

type ErrorLine struct {
	Message string
}

func (event *ChatEvent) String() string {
	return fmt.Sprintf("%v %v", event.Time.Format(time.DateTime), event.Contents)
}

func (event *Login) ImplementsChatContent() {}
func (event *Login) String() string {
	return fmt.Sprintf("**** Logged in as '%v' ****", event.Name)
}

func (event *SkillUse) ImplementsChatContent() {}
func (event *SkillUse) String() string {
	var fatality string
	if event.Fatality {
		fatality = " FATAL!"
	}

	var evaded string
	if event.Evaded {
		evaded = " EVADED!"
	}

	var crit string
	if event.Crit {
		crit = " CRIT!"
	}

	if event.Damage != nil {
		return fmt.Sprintf("'%v' used '%v' on '%v'%v%v%v with damage %v",
			event.Subject, event.Skill, event.Victim, crit, evaded, fatality, event.Damage)
	} else {
		return fmt.Sprintf("'%v' used '%v' on '%v'%v%v%v", event.Subject, event.Skill, event.Victim, crit, evaded, fatality)
	}
}

func (event Vitals) String() string {
	return utils.FormatDamageLabel(event.Health, event.Armor, event.Power)
}
func (event Vitals) Add(other Vitals) Vitals {
	return Vitals{
		Health: event.Health + other.Health,
		Armor:  event.Armor + other.Armor,
		Power:  event.Power + other.Power,
	}
}
func (event Vitals) Abs() Vitals {
	return Vitals{
		Health: utils.AbsInt(event.Health),
		Armor:  utils.AbsInt(event.Armor),
		Power:  utils.AbsInt(event.Power),
	}
}
func (event Vitals) Total() int {
	return event.Health + event.Armor + event.Power
}

// StringCL Format vitals as conditionally long number
// StringCL Format vitals as conditionally long numbers
func (event Vitals) StringCL(long bool) string {
	return utils.ConditionalDamageLabel(event.Health, event.Armor, event.Power, long)
}

func (event *Recovered) ImplementsChatContent() {}
func (event *Recovered) String() string {
	return fmt.Sprintf("'%v' recovered %v", event.Subject, event.Healed)
}

func (event *IndirectDamage) ImplementsChatContent() {}
func (event *IndirectDamage) String() string {
	return fmt.Sprintf("'%v' suffered indirect Damage of %v", event.Subject, event.Damage)
}

func (event *XPGained) ImplementsChatContent() {}
func (event *XPGained) String() string {
	return fmt.Sprintf("%v XP gained in '%v'", event.XP, event.Skill)
}

func (event *XPGainedLeveledUp) ImplementsChatContent() {}
func (event *XPGainedLeveledUp) String() string {
	return fmt.Sprintf("%v XP gained in '%v' and leveled up to '%v'", event.XP, event.Skill, event.Level)
}

func (event *FoundCoins) ImplementsChatContent() {}
func (event *FoundCoins) String() string {
	return fmt.Sprintf("Found %v coins", event.Coins)
}

func (event *ReceivedCoins) ImplementsChatContent() {}
func (event *ReceivedCoins) String() string {
	return fmt.Sprintf("Received %v coins", event.Coins)
}

func (e *ErrorLine) ImplementsChatContent() {}
func (e *ErrorLine) String() string {
	return fmt.Sprintf("[ERROR]: %v", e.Message)
}
