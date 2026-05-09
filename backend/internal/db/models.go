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

type Session struct {
	ID                string     `gorm:"primaryKey;type:text"`
	UserID            string     `gorm:"type:text"`
	ScenarioID        string     `gorm:"not null;index"`
	Status            string     `gorm:"not null"`
	CurrentSequenceNo int        `gorm:"not null;default:0"`
	Mode              string     `gorm:"not null"`
	StartedAt         time.Time
	CompletedAt       *time.Time

	Scenario Scenario       `gorm:"foreignKey:ScenarioID"`
	Events   []SessionEvent `gorm:"foreignKey:SessionID"`
}

type SessionEvent struct {
	ID           string    `gorm:"primaryKey;type:text"`
	SessionID    string    `gorm:"not null;index"`
	SequenceNo   int       `gorm:"not null"`
	EventType    string    `gorm:"not null"`
	Actor        string
	Text         string
	MetadataJSON string    `gorm:"column:metadata_json"`
	CreatedAt    time.Time
}

type TraineeAction struct {
	ID                        string    `gorm:"primaryKey;type:text"`
	SessionID                 string    `gorm:"not null;index"`
	ScenarioLineID            *string   `gorm:"type:text;index"`
	ActionType                string    `gorm:"not null"`
	RawText                   string    `gorm:"not null"`
	NormalizedObjectionTypeID *string   `gorm:"type:text"`
	CreatedAt                 time.Time

	Session                 Session       `gorm:"foreignKey:SessionID"`
	ScenarioLine            ScenarioLine  `gorm:"foreignKey:ScenarioLineID"`
	NormalizedObjectionType ObjectionType `gorm:"foreignKey:NormalizedObjectionTypeID"`
}

type ActionEvaluation struct {
	ID                         string    `gorm:"primaryKey;type:text"`
	TraineeActionID            string    `gorm:"not null;index"`
	MatchedOpportunityID       *string   `gorm:"type:text;index"`
	NormalizedObjectionTypeID  *string   `gorm:"type:text"`
	Valid                      bool      `gorm:"not null;default:false"`
	Timely                     bool      `gorm:"not null;default:false"`
	Ruling                     string    `gorm:"not null"`
	LegalAccuracyScore         float64   `gorm:"not null;default:0"`
	PhrasingScore              float64   `gorm:"not null;default:0"`
	StrategyScore              float64   `gorm:"not null;default:0"`
	Feedback                   string    `gorm:"not null"`
	CreatedAt                  time.Time

	TraineeAction           TraineeAction          `gorm:"foreignKey:TraineeActionID"`
	MatchedOpportunity      ObjectionOpportunity   `gorm:"foreignKey:MatchedOpportunityID"`
	NormalizedObjectionType ObjectionType          `gorm:"foreignKey:NormalizedObjectionTypeID"`
}

type SessionScore struct {
	ID                     string    `gorm:"primaryKey;type:text"`
	SessionID              string    `gorm:"not null;uniqueIndex"`
	EvaluatedActionCount   int       `gorm:"not null;default:0"`
	TotalOpportunityCount  int       `gorm:"not null;default:0"`
	MatchedOpportunityCount int      `gorm:"not null;default:0"`
	MissedOpportunityCount int       `gorm:"not null;default:0"`
	FalsePositiveCount     int       `gorm:"not null;default:0"`

	SpottingAccuracy float64 `gorm:"not null;default:0"`
	LegalAccuracy    float64 `gorm:"not null;default:0"`
	Timeliness        float64 `gorm:"not null;default:0"`
	Phrasing          float64 `gorm:"not null;default:0"`
	Strategy          float64 `gorm:"not null;default:0"`
	ResponseQuality   float64 `gorm:"not null;default:0"`
	OverallScore      float64 `gorm:"not null;default:0"`

	IsFinal   bool      `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Session Session `gorm:"foreignKey:SessionID"`
}