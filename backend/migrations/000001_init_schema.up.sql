CREATE TABLE IF NOT EXISTS scenarios (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    jurisdiction TEXT NOT NULL,
    practice_area TEXT NOT NULL,
    hearing_type TEXT NOT NULL,
    difficulty TEXT NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
    status TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scenario_actors (
    id TEXT PRIMARY KEY,
    scenario_id TEXT NOT NULL,
    name TEXT NOT NULL,
    actor_type TEXT NOT NULL CHECK (
        actor_type IN (
            'judge',
            'witness',
            'opposing_counsel',
            'trainee_counsel'
        )
    ),
    persona TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_scenario_actors_scenario
        FOREIGN KEY (scenario_id)
        REFERENCES scenarios(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS scenario_lines (
    id TEXT PRIMARY KEY,
    scenario_id TEXT NOT NULL,
    sequence_no INTEGER NOT NULL,
    speaker_type TEXT NOT NULL CHECK (
        speaker_type IN (
            'judge',
            'witness',
            'opposing_counsel',
            'trainee_counsel',
            'coach',
            'system'
        )
    ),
    speaker_name TEXT,
    line_text TEXT NOT NULL,
    line_kind TEXT NOT NULL CHECK (
        line_kind IN (
            'question',
            'answer',
            'argument',
            'ruling',
            'instruction'
        )
    ),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_scenario_lines_scenario
        FOREIGN KEY (scenario_id)
        REFERENCES scenarios(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_scenario_lines_sequence
        UNIQUE (scenario_id, sequence_no)
);

CREATE TABLE IF NOT EXISTS objection_types (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    default_phrase TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rule_refs (
    id TEXT PRIMARY KEY,
    jurisdiction TEXT NOT NULL,
    rule_code TEXT NOT NULL,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    source_text TEXT,
    citation TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_rule_refs_jurisdiction_rule
        UNIQUE (jurisdiction, rule_code)
);

CREATE TABLE IF NOT EXISTS objection_opportunities (
    id TEXT PRIMARY KEY,
    scenario_line_id TEXT NOT NULL,
    objection_type_id TEXT NOT NULL,
    strength TEXT NOT NULL CHECK (strength IN ('weak', 'moderate', 'strong')),
    timing_window TEXT NOT NULL CHECK (
        timing_window IN (
            'after_question',
            'after_answer',
            'before_answer'
        )
    ),
    explanation TEXT NOT NULL,
    expected_phrase TEXT,
    is_primary INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_objection_opportunities_line
        FOREIGN KEY (scenario_line_id)
        REFERENCES scenario_lines(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_objection_opportunities_type
        FOREIGN KEY (objection_type_id)
        REFERENCES objection_types(id)
);

CREATE TABLE IF NOT EXISTS opportunity_rule_refs (
    opportunity_id TEXT NOT NULL,
    rule_ref_id TEXT NOT NULL,

    PRIMARY KEY (opportunity_id, rule_ref_id),

    CONSTRAINT fk_opportunity_rule_refs_opportunity
        FOREIGN KEY (opportunity_id)
        REFERENCES objection_opportunities(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_opportunity_rule_refs_rule
        FOREIGN KEY (rule_ref_id)
        REFERENCES rule_refs(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scenarios_status
    ON scenarios(status);

CREATE INDEX IF NOT EXISTS idx_scenarios_difficulty
    ON scenarios(difficulty);

CREATE INDEX IF NOT EXISTS idx_scenario_lines_scenario_sequence
    ON scenario_lines(scenario_id, sequence_no);

CREATE INDEX IF NOT EXISTS idx_objection_opportunities_line
    ON objection_opportunities(scenario_line_id);