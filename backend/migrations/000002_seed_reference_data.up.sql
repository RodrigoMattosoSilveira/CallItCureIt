INSERT INTO objection_types (
    id,
    code,
    name,
    description,
    default_phrase
) VALUES
(
    'obj-hearsay',
    'hearsay',
    'Hearsay',
    'An out-of-court statement offered to prove the truth of the matter asserted.',
    'Objection, hearsay.'
),
(
    'obj-relevance',
    'relevance',
    'Relevance',
    'Evidence that does not tend to make a fact of consequence more or less probable.',
    'Objection, relevance.'
),
(
    'obj-foundation',
    'foundation',
    'Lack of Foundation',
    'The proponent has not established the necessary preliminary facts for the evidence or testimony.',
    'Objection, lack of foundation.'
),
(
    'obj-leading',
    'leading',
    'Leading Question',
    'A question that suggests the answer, generally improper on direct examination.',
    'Objection, leading.'
),
(
    'obj-speculation',
    'speculation',
    'Speculation',
    'The question asks the witness to guess or testify beyond personal knowledge.',
    'Objection, calls for speculation.'
),
(
    'obj-asked-answered',
    'asked_and_answered',
    'Asked and Answered',
    'The question has already been asked and answered.',
    'Objection, asked and answered.'
),
(
    'obj-argumentative',
    'argumentative',
    'Argumentative',
    'The question argues with the witness rather than seeking admissible testimony.',
    'Objection, argumentative.'
),
(
    'obj-compound',
    'compound',
    'Compound Question',
    'The question combines multiple questions in a way that may confuse the witness or record.',
    'Objection, compound.'
);

INSERT INTO rule_refs (
    id,
    jurisdiction,
    rule_code,
    title,
    summary,
    source_text,
    citation
) VALUES
(
    'rule-fre-401',
    'federal',
    'FRE 401',
    'Test for Relevant Evidence',
    'Evidence is relevant if it tends to make a fact more or less probable and the fact is of consequence in determining the action.',
    NULL,
    'Federal Rule of Evidence 401'
),
(
    'rule-fre-402',
    'federal',
    'FRE 402',
    'General Admissibility of Relevant Evidence',
    'Relevant evidence is generally admissible unless otherwise excluded; irrelevant evidence is not admissible.',
    NULL,
    'Federal Rule of Evidence 402'
),
(
    'rule-fre-602',
    'federal',
    'FRE 602',
    'Need for Personal Knowledge',
    'A witness may testify only if evidence supports a finding that the witness has personal knowledge of the matter.',
    NULL,
    'Federal Rule of Evidence 602'
),
(
    'rule-fre-611',
    'federal',
    'FRE 611',
    'Mode and Order of Examining Witnesses and Presenting Evidence',
    'The court should exercise reasonable control over examining witnesses and presenting evidence. Leading questions are generally not allowed on direct examination except as necessary to develop testimony.',
    NULL,
    'Federal Rule of Evidence 611'
),
(
    'rule-fre-801',
    'federal',
    'FRE 801',
    'Definitions That Apply to Hearsay',
    'Hearsay depends on a statement, the declarant, and whether the statement is offered for the truth of the matter asserted.',
    NULL,
    'Federal Rule of Evidence 801'
),
(
    'rule-fre-802',
    'federal',
    'FRE 802',
    'The Rule Against Hearsay',
    'Hearsay is not admissible unless allowed by federal statute, the Federal Rules of Evidence, or other rules prescribed by the Supreme Court.',
    NULL,
    'Federal Rule of Evidence 802'
);

INSERT INTO scenarios (
    id,
    title,
    description,
    jurisdiction,
    practice_area,
    hearing_type,
    difficulty,
    status
) VALUES
(
    'scenario-hearsay-001',
    'Basic Hearsay on Direct Examination',
    'A beginner scenario focused on identifying a hearsay objection during witness testimony.',
    'federal',
    'civil',
    'trial_direct_examination',
    'beginner',
    'published'
);

INSERT INTO scenario_actors (
    id,
    scenario_id,
    name,
    actor_type,
    persona
) VALUES
(
    'actor-judge-001',
    'scenario-hearsay-001',
    'Judge Carter',
    'judge',
    'Firm but instructional. Gives short rulings.'
),
(
    'actor-opposing-001',
    'scenario-hearsay-001',
    'Ms. Daniels',
    'opposing_counsel',
    'Experienced civil litigator conducting direct examination.'
),
(
    'actor-witness-001',
    'scenario-hearsay-001',
    'John Miller',
    'witness',
    'Neighbor of one of the parties.'
),
(
    'actor-trainee-001',
    'scenario-hearsay-001',
    'Trainee Counsel',
    'trainee_counsel',
    'The lawyer being trained.'
);

INSERT INTO scenario_lines (
    id,
    scenario_id,
    sequence_no,
    speaker_type,
    speaker_name,
    line_text,
    line_kind
) VALUES
(
    'line-hearsay-001',
    'scenario-hearsay-001',
    1,
    'opposing_counsel',
    'Ms. Daniels',
    'Mr. Miller, where were you on the evening of March 12?',
    'question'
),
(
    'line-hearsay-002',
    'scenario-hearsay-001',
    2,
    'witness',
    'John Miller',
    'I was standing on my front porch.',
    'answer'
),
(
    'line-hearsay-003',
    'scenario-hearsay-001',
    3,
    'opposing_counsel',
    'Ms. Daniels',
    'Did you speak with your neighbor that evening?',
    'question'
),
(
    'line-hearsay-004',
    'scenario-hearsay-001',
    4,
    'witness',
    'John Miller',
    'Yes. She told me that the defendant admitted he caused the accident.',
    'answer'
),
(
    'line-hearsay-005',
    'scenario-hearsay-001',
    5,
    'opposing_counsel',
    'Ms. Daniels',
    'What did you do after that conversation?',
    'question'
),
(
    'line-hearsay-006',
    'scenario-hearsay-001',
    6,
    'witness',
    'John Miller',
    'I called the police.',
    'answer'
);

INSERT INTO objection_opportunities (
    id,
    scenario_line_id,
    objection_type_id,
    strength,
    timing_window,
    explanation,
    expected_phrase,
    is_primary
) VALUES
(
    'opp-hearsay-001',
    'line-hearsay-004',
    'obj-hearsay',
    'strong',
    'after_answer',
    'The witness is repeating an out-of-court statement from the neighbor, offered to prove that the defendant admitted causing the accident.',
    'Objection, hearsay.',
    1
);

INSERT INTO opportunity_rule_refs (
    opportunity_id,
    rule_ref_id
) VALUES
(
    'opp-hearsay-001',
    'rule-fre-801'
),
(
    'opp-hearsay-001',
    'rule-fre-802'
);