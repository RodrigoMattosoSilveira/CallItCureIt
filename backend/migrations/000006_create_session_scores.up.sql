DROP TABLE IF EXISTS session_scores;

CREATE TABLE session_scores (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL UNIQUE,

    evaluated_action_count INTEGER NOT NULL DEFAULT 0,
    total_opportunity_count INTEGER NOT NULL DEFAULT 0,
    matched_opportunity_count INTEGER NOT NULL DEFAULT 0,
    missed_opportunity_count INTEGER NOT NULL DEFAULT 0,
    false_positive_count INTEGER NOT NULL DEFAULT 0,

    spotting_accuracy REAL NOT NULL DEFAULT 0,
    legal_accuracy REAL NOT NULL DEFAULT 0,
    timeliness REAL NOT NULL DEFAULT 0,
    phrasing REAL NOT NULL DEFAULT 0,
    strategy REAL NOT NULL DEFAULT 0,
    response_quality REAL NOT NULL DEFAULT 0,
    overall_score REAL NOT NULL DEFAULT 0,

    is_final INTEGER NOT NULL DEFAULT 0,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_session_scores_session
        FOREIGN KEY (session_id)
        REFERENCES sessions(id)
        ON DELETE CASCADE
);