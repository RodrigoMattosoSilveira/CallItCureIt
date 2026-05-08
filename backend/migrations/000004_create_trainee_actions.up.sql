CREATE TABLE IF NOT EXISTS trainee_actions (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    scenario_line_id TEXT,
    action_type TEXT NOT NULL CHECK (
        action_type IN (
            'object',
            'respond',
            'pass'
        )
    ),
    raw_text TEXT NOT NULL,
    normalized_objection_type_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_trainee_actions_session
        FOREIGN KEY (session_id)
        REFERENCES sessions(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_trainee_actions_scenario_line
        FOREIGN KEY (scenario_line_id)
        REFERENCES scenario_lines(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_trainee_actions_normalized_objection_type
        FOREIGN KEY (normalized_objection_type_id)
        REFERENCES objection_types(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_trainee_actions_session_id
    ON trainee_actions(session_id);

CREATE INDEX IF NOT EXISTS idx_trainee_actions_scenario_line_id
    ON trainee_actions(scenario_line_id);