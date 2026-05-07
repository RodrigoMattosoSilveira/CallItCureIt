DELETE FROM opportunity_rule_refs
WHERE opportunity_id = 'opp-hearsay-001';

DELETE FROM objection_opportunities
WHERE id = 'opp-hearsay-001';

DELETE FROM scenario_lines
WHERE scenario_id = 'scenario-hearsay-001';

DELETE FROM scenario_actors
WHERE scenario_id = 'scenario-hearsay-001';

DELETE FROM scenarios
WHERE id = 'scenario-hearsay-001';

DELETE FROM rule_refs
WHERE id IN (
    'rule-fre-401',
    'rule-fre-402',
    'rule-fre-602',
    'rule-fre-611',
    'rule-fre-801',
    'rule-fre-802'
);

DELETE FROM objection_types
WHERE id IN (
    'obj-hearsay',
    'obj-relevance',
    'obj-foundation',
    'obj-leading',
    'obj-speculation',
    'obj-asked-answered',
    'obj-argumentative',
    'obj-compound'
);