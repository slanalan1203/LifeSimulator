package Colonist

const (
	MaxSkillLevel = 5
	XPerLevel     = 50
)

type Skill struct {
	XP int
}

func (s Skill) Level() int {
	return min(MaxSkillLevel, 1+s.XP/XPerLevel)
}

func (s *Skill) AddXp() {
	maxXp := (MaxSkillLevel - 1) * XPerLevel
	s.XP = min(maxXp, s.XP+1)
}
