package db

import "time"

type Scenario struct {
	ID           string    `gorm:"primaryKey;type:text"`
	Title        string    `gorm:"not null"`
	Description  string
	Jurisdiction string    `gorm:"not null"`
	PracticeArea string    `gorm:"not null"`
	HearingType  string    `gorm:"not null"`
	Difficulty   string    `gorm:"not null"`
	Status       string    `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Actors []ScenarioActor `gorm:"foreignKey:ScenarioID"`
	Lines  []ScenarioLine  `gorm:"foreignKey:ScenarioID"`
}

type ScenarioActor struct {
	ID         string    `gorm:"primaryKey;type:text"`
	ScenarioID string    `gorm:"not null;index"`
	Name       string    `gorm:"not null"`
	ActorType  string    `gorm:"not null"`
	Persona    string
	CreatedAt  time.Time
}

type ScenarioLine struct {
	ID          string    `gorm:"primaryKey;type:text"`
	ScenarioID  string    `gorm:"not null;index"`
	SequenceNo  int       `gorm:"not null"`
	SpeakerType string    `gorm:"not null"`
	SpeakerName string
	LineText    string    `gorm:"not null"`
	LineKind    string    `gorm:"not null"`
	CreatedAt   time.Time

	Opportunities []ObjectionOpportunity `gorm:"foreignKey:ScenarioLineID"`
}

type ObjectionType struct {
	ID            string    `gorm:"primaryKey;type:text"`
	Code          string    `gorm:"not null;uniqueIndex"`
	Name          string    `gorm:"not null"`
	Description   string    `gorm:"not null"`
	DefaultPhrase string    `gorm:"not null"`
	CreatedAt     time.Time
}

type RuleRef struct {
	ID           string    `gorm:"primaryKey;type:text"`
	Jurisdiction string    `gorm:"not null"`
	RuleCode     string    `gorm:"not null"`
	Title        string    `gorm:"not null"`
	Summary      string    `gorm:"not null"`
	SourceText   string
	Citation     string    `gorm:"not null"`
	CreatedAt    time.Time
}

type ObjectionOpportunity struct {
	ID                string    `gorm:"primaryKey;type:text"`
	ScenarioLineID    string    `gorm:"not null;index"`
	ObjectionTypeID    string    `gorm:"not null;index"`
	Strength          string    `gorm:"not null"`
	TimingWindow      string    `gorm:"not null"`
	Explanation       string    `gorm:"not null"`
	ExpectedPhrase    string
	IsPrimary         bool
	CreatedAt         time.Time

	ObjectionType ObjectionType `gorm:"foreignKey:ObjectionTypeID"`
	RuleRefs      []RuleRef     `gorm:"many2many:opportunity_rule_refs;foreignKey:ID;joinForeignKey:opportunity_id;References:ID;joinReferences:rule_ref_id"`
}