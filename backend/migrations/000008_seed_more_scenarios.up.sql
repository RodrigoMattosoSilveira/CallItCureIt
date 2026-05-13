-- Additional beginner scenarios for broader objection coverage.

-- ============================================================
-- 1. Leading question
-- ============================================================

INSERT OR IGNORE INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES (
    'scenario-leading-001',
    'Basic Leading Question on Direct Examination',
    'A beginner scenario where opposing counsel asks a leading question during direct examination.',
    'federal',
    'civil',
    'trial_direct_examination',
    'beginner',
    'published'
);

INSERT OR IGNORE INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-leading-001',
    'scenario-leading-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'Mr. Miller, you were standing on your front porch that evening, correct?',
    'question'
),
(
    'line-leading-002',
    'scenario-leading-001',
    2,
    'witness',
    'John Miller',
    'Yes.',
    'answer'
),
(
    'line-leading-003',
    'scenario-leading-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'And you clearly saw the defendant run the red light, correct?',
    'question'
);

INSERT OR IGNORE INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES (
    'opp-leading-001',
    'line-leading-003',
    'obj-leading',
    'strong',
    'after_question',
    'The question suggests the desired answer and is being asked on direct examination.',
    'Objection, leading.',
    1
);

-- ============================================================
-- 2. Lack of foundation
-- ============================================================

INSERT OR IGNORE INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES (
    'scenario-foundation-001',
    'Basic Lack of Foundation',
    'A beginner scenario where the witness identifies something without first establishing personal knowledge or basis.',
    'federal',
    'civil',
    'trial_direct_examination',
    'beginner',
    'published'
);

INSERT OR IGNORE INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-foundation-001',
    'scenario-foundation-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'Do you recognize this document?',
    'question'
),
(
    'line-foundation-002',
    'scenario-foundation-001',
    2,
    'witness',
    'John Miller',
    'Yes, that is the repair estimate for the defendant’s car.',
    'answer'
),
(
    'line-foundation-003',
    'scenario-foundation-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'How do you know that?',
    'question'
);

INSERT OR IGNORE INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES (
    'opp-foundation-001',
    'line-foundation-002',
    'obj-foundation',
    'strong',
    'after_answer',
    'The witness identifies the document and its meaning before counsel has established how the witness recognizes it or has personal knowledge of it.',
    'Objection, lack of foundation.',
    1
);

-- ============================================================
-- 3. Speculation
-- ============================================================

INSERT OR IGNORE INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES (
    'scenario-speculation-001',
    'Basic Speculation',
    'A beginner scenario where counsel asks the witness to guess about another person’s intent.',
    'federal',
    'civil',
    'trial_direct_examination',
    'beginner',
    'published'
);

INSERT OR IGNORE INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-speculation-001',
    'scenario-speculation-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'What did you see after the light turned red?',
    'question'
),
(
    'line-speculation-002',
    'scenario-speculation-001',
    2,
    'witness',
    'John Miller',
    'The defendant’s car entered the intersection.',
    'answer'
),
(
    'line-speculation-003',
    'scenario-speculation-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'Why do you think the defendant decided to speed through the intersection?',
    'question'
);

INSERT OR IGNORE INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES (
    'opp-speculation-001',
    'line-speculation-003',
    'obj-speculation',
    'strong',
    'after_question',
    'The question asks the witness to guess about the defendant’s intent rather than testify from personal knowledge.',
    'Objection, calls for speculation.',
    1
);

-- ============================================================
-- 4. Relevance
-- ============================================================

INSERT OR IGNORE INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES (
    'scenario-relevance-001',
    'Basic Relevance',
    'A beginner scenario where counsel asks about a fact that does not tend to prove or disprove a material issue.',
    'federal',
    'civil',
    'trial_direct_examination',
    'beginner',
    'published'
);

INSERT OR IGNORE INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-relevance-001',
    'scenario-relevance-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'You witnessed the collision on March 12?',
    'question'
),
(
    'line-relevance-002',
    'scenario-relevance-001',
    2,
    'witness',
    'John Miller',
    'Yes.',
    'answer'
),
(
    'line-relevance-003',
    'scenario-relevance-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'What kind of music does the defendant usually listen to?',
    'question'
);

INSERT OR IGNORE INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES (
    'opp-relevance-001',
    'line-relevance-003',
    'obj-relevance',
    'strong',
    'after_question',
    'The defendant’s usual music preference does not appear to make any material fact about the collision more or less probable.',
    'Objection, relevance.',
    1
);

-- ============================================================
-- 5. Compound question
-- ============================================================

INSERT OR IGNORE INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES (
    'scenario-compound-001',
    'Basic Compound Question',
    'A beginner scenario where counsel asks two questions at once.',
    'federal',
    'civil',
    'trial_direct_examination',
    'beginner',
    'published'
);

INSERT OR IGNORE INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-compound-001',
    'scenario-compound-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'Where were you standing when the accident happened?',
    'question'
),
(
    'line-compound-002',
    'scenario-compound-001',
    2,
    'witness',
    'John Miller',
    'On my front porch.',
    'answer'
),
(
    'line-compound-003',
    'scenario-compound-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'Did you see the defendant enter the intersection and did you call the police immediately afterward?',
    'question'
);

INSERT OR IGNORE INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES (
    'opp-compound-001',
    'line-compound-003',
    'obj-compound',
    'strong',
    'after_question',
    'The question combines two separate factual questions, which can confuse the witness and the record.',
    'Objection, compound.',
    1
);

-- ============================================================
-- 6. Asked and answered
-- ============================================================

INSERT OR IGNORE INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES (
    'scenario-asked-answered-001',
    'Basic Asked and Answered',
    'A beginner scenario where counsel repeats a question the witness already answered.',
    'federal',
    'civil',
    'trial_direct_examination',
    'beginner',
    'published'
);

INSERT OR IGNORE INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-asked-answered-001',
    'scenario-asked-answered-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'Were you standing on your front porch when the accident happened?',
    'question'
),
(
    'line-asked-answered-002',
    'scenario-asked-answered-001',
    2,
    'witness',
    'John Miller',
    'Yes, I was standing on my front porch.',
    'answer'
),
(
    'line-asked-answered-003',
    'scenario-asked-answered-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'So you were standing on your front porch when the accident happened?',
    'question'
);

INSERT OR IGNORE INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES (
    'opp-asked-answered-001',
    'line-asked-answered-003',
    'obj-asked-answered',
    'strong',
    'after_question',
    'Counsel is repeating a question the witness has already answered.',
    'Objection, asked and answered.',
    1
);

-- ============================================================
-- 7. Argumentative
-- ============================================================

INSERT OR IGNORE INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES (
    'scenario-argumentative-001',
    'Basic Argumentative Question',
    'A beginner scenario where counsel argues with the witness instead of asking a proper question.',
    'federal',
    'civil',
    'trial_cross_examination',
    'beginner',
    'published'
);

INSERT OR IGNORE INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-argumentative-001',
    'scenario-argumentative-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'You testified that you were watching the intersection?',
    'question'
),
(
    'line-argumentative-002',
    'scenario-argumentative-001',
    2,
    'witness',
    'John Miller',
    'Yes.',
    'answer'
),
(
    'line-argumentative-003',
    'scenario-argumentative-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'Isn’t it true that you are just making this story up to help your friend?',
    'question'
);

INSERT OR IGNORE INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES (
    'opp-argumentative-001',
    'line-argumentative-003',
    'obj-argumentative',
    'strong',
    'after_question',
    'The question is argumentative because counsel is attacking the witness with an accusation rather than asking a neutral factual question.',
    'Objection, argumentative.',
    1
);