CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    scenario_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'abandoned')),
    current_sequence_no INTEGER NOT NULL DEFAULT 0,
    mode TEXT NOT NULL CHECK (
        mode IN (
            'spot_objection',
            'respond_to_objection',
            'full_simulation'
        )
    ),
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,

    CONSTRAINT fk_sessions_scenario
        FOREIGN KEY (scenario_id)
        REFERENCES scenarios(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS session_events (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    sequence_no INTEGER NOT NULL,
    event_type TEXT NOT NULL CHECK (
        event_type IN (
            'system_line',
            'trainee_objection',
            'trainee_response',
            'judge_ruling',
            'coach_feedback',
            'missed_opportunity'
        )
    ),
    actor TEXT,
    text TEXT,
    metadata_json TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_session_events_session
        FOREIGN KEY (session_id)
        REFERENCES sessions(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_scenario_id
    ON sessions(scenario_id);

CREATE INDEX IF NOT EXISTS idx_sessions_status
    ON sessions(status);

CREATE INDEX IF NOT EXISTS idx_session_events_session_sequence
    ON session_events(session_id, sequence_no);