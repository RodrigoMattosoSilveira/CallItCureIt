CREATE TABLE IF NOT EXISTS action_evaluations (
    id TEXT PRIMARY KEY,
    trainee_action_id TEXT NOT NULL,
    matched_opportunity_id TEXT,
    normalized_objection_type_id TEXT,
    valid INTEGER NOT NULL DEFAULT 0,
    timely INTEGER NOT NULL DEFAULT 0,
    ruling TEXT NOT NULL CHECK (
        ruling IN (
            'sustained',
            'overruled',
            'no_ruling'
        )
    ),
    legal_accuracy_score REAL NOT NULL DEFAULT 0,
    phrasing_score REAL NOT NULL DEFAULT 0,
    strategy_score REAL NOT NULL DEFAULT 0,
    feedback TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_action_evaluations_action
        FOREIGN KEY (trainee_action_id)
        REFERENCES trainee_actions(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_action_evaluations_opportunity
        FOREIGN KEY (matched_opportunity_id)
        REFERENCES objection_opportunities(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_action_evaluations_objection_type
        FOREIGN KEY (normalized_objection_type_id)
        REFERENCES objection_types(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_action_evaluations_action
    ON action_evaluations(trainee_action_id);

CREATE INDEX IF NOT EXISTS idx_action_evaluations_matched_opportunity
    ON action_evaluations(matched_opportunity_id);