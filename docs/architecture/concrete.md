# 

[**1\. MVP scope	3**](#1.-mvp-scope)

[**2\. Service boundaries	3**](#2.-service-boundaries)

[A. Frontend	3](#a.-frontend)

[**3\. Backend services	4**](#3.-backend-services)

[3.1 Auth Service	4](#3.1-auth-service)

[3.2 Scenario Service	4](#3.2-scenario-service)

[3.3 Session Orchestrator	5](#3.3-session-orchestrator)

[3.4 Objection Rules Service	5](#3.4-objection-rules-service)

[3.5 LLM Simulation Service	6](#3.5-llm-simulation-service)

[3.6 Scoring Service	6](#3.6-scoring-service)

[3.7 Feedback/Debrief Service	7](#3.7-feedback/debrief-service)

[**4\. Suggested MVP database schema	7**](#4.-suggested-mvp-database-schema)

[users	7](#users)

[scenarios	7](#scenarios)

[scenario\_actors	8](#scenario_actors)

[scenario\_lines	8](#scenario_lines)

[objection\_types	9](#objection_types)

[rule\_refs	9](#rule_refs)

[objection\_opportunities	9](#objection_opportunities)

[opportunity\_rule\_refs	10](#opportunity_rule_refs)

[sessions	10](#sessions)

[session\_events	10](#session_events)

[trainee\_actions	11](#trainee_actions)

[action\_evaluations	11](#action_evaluations)

[session\_scores	12](#session_scores)

[**5\. Sample event flow	12**](#5.-sample-event-flow)

[Flow: trainee spots a hearsay objection	12](#flow:-trainee-spots-a-hearsay-objection)

[Example transcript	13](#example-transcript)

[**6\. API endpoints	13**](#6.-api-endpoints)

[Auth	13](#auth)

[Scenarios	13](#scenarios-1)

[Sessions	14](#sessions-1)

[Example create session	14](#example-create-session)

[Example trainee action	14](#example-trainee-action)

[Example response	14](#example-response)

[**7\. Internal service flow	15**](#7.-internal-service-flow)

[**8\. MVP rule evaluation logic	16**](#8.-mvp-rule-evaluation-logic)

[**9\. Suggested MVP folder structure	17**](#9.-suggested-mvp-folder-structure)

[**10\. MVP build order	18**](#10.-mvp-build-order)

# 

# **1\. MVP scope** {#1.-mvp-scope}

Start narrow:

* **Jurisdiction:** Federal Rules of Evidence  
* **Training type:** trial objection practice  
* **Mode:** text-first courtroom simulation  
* **Actors:** trainee, opposing counsel, witness, judge, coach  
* **Core skill:** spot objection opportunities and respond to objections  
* **Not included in MVP:** real-time voice, multi-jurisdiction support, real case upload, CLE administration

# **2\. Service boundaries** {#2.-service-boundaries}

```bash
React Frontend
  |
  v
API Gateway / Backend
  |
  +-- Auth Service
  +-- Scenario Service
  +-- Session Orchestrator
  +-- Objection Rules Service
  +-- LLM Simulation Service
  +-- Scoring Service
  +-- Feedback/Debrief Service
  |
  v
PostgreSQL
Vector / Document Index
Object Storage later
```

## **A. Frontend** {#a.-frontend}

Responsibilities:

* scenario selection  
* courtroom simulation UI  
* transcript display  
* trainee input box  
* objection timer  
* judge ruling panel  
* score/debrief screen

Main screens:
```bash
/login
/dashboard
/scenarios
/sessions/:id/play
/sessions/:id/debrief
/admin/scenarios
```

# **3\. Backend services** {#3.-backend-services}

## **3.1 Auth Service** {#3.1-auth-service}

Handles:

* users  
* roles  
* organizations  
* instructor/admin access

Roles:
```bash
student
lawyer
instructor
admin
```

## **3.2 Scenario Service** {#3.2-scenario-service}

Owns scenario content.

It stores:

* scenario metadata  
* characters  
* scripted lines  
* expected objection opportunities  
* model answers  
* difficulty level

Example responsibility:

Load next scripted witness/counsel line.

Return expected objection windows for this line.

## **3.3 Session Orchestrator** {#3.3-session-orchestrator}

This is the heart of the MVP.

Responsibilities:

* starts a training session  
* tracks current transcript position  
* receives trainee actions  
* calls rules engine  
* calls judge simulator  
* stores events  
* advances the session

Do not put legal logic directly in controllers. The orchestrator coordinates; other services decide.

## **3.4 Objection Rules Service** {#3.4-objection-rules-service}

Determines whether a trainee action is legally plausible.

Responsibilities:

* classify objection  
* compare against expected objection opportunities  
* determine timeliness  
* determine if response cures the problem  
* return structured result

Example output:
```json
{
 "valid": true,
 "objection_type": "hearsay",
 "timely": true,
 "strength": "strong",
 "expected_opportunity_id": "opp_123",
 "rule_refs": ["FRE 801", "FRE 802"],
 "notes": "The witness is repeating an out-of-court statement offered for its truth."
}
```

## **3.5 LLM Simulation Service** {#3.5-llm-simulation-service}

Use the LLM only for controlled tasks:

* generate natural witness/counsel lines from approved scenario state  
* simulate judge language  
* provide coaching explanations  
* paraphrase feedback

***Do not let the LLM be the sole authority on whether the objection is correct***.

## **3.6 Scoring Service** {#3.6-scoring-service}

Computes:

* spotted valid opportunities  
* missed opportunities  
* false positives  
* wrong objection grounds  
* late objections  
* quality of phrasing  
* response effectiveness

Example scoring dimensions:
```bash
spotting_accuracy
legal_accuracy
timeliness
strategic_judgment
oral_concision
response_quality
```

## **3.7 Feedback/Debrief Service** {#3.7-feedback/debrief-service}

Builds the post-session review.

It should show:

* transcript  
* missed objections  
* correct objections  
* weak objections  
* better phrasing  
* rule explanation  
* score trend

# **4\. Suggested MVP database schema** {#4.-suggested-mvp-database-schema}

PostgreSQL is enough.

## **users** {#users}
```sql
CREATE TABLE users (
   id UUID PRIMARY KEY,
   email TEXT NOT NULL UNIQUE,
   name TEXT NOT NULL,
   role TEXT NOT NULL CHECK (role IN ('student', 'lawyer', 'instructor', 'admin')),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## **scenarios** {#scenarios}
```sql
CREATE TABLE scenarios (
   id UUID PRIMARY KEY,
   title TEXT NOT NULL,
   description TEXT,
   jurisdiction TEXT NOT NULL,
   practice_area TEXT NOT NULL,
   hearing_type TEXT NOT NULL,
   difficulty TEXT NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
   status TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
   created_by UUID REFERENCES users(id),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## **scenario\_actors** {#scenario_actors}
```sql
CREATE TABLE scenario_actors (
   id UUID PRIMARY KEY,
   scenario_id UUID NOT NULL REFERENCES scenarios(id),
   name TEXT NOT NULL,
   actor_type TEXT NOT NULL CHECK (actor_type IN ('judge', 'witness', 'opposing_counsel', 'trainee_counsel')),
   persona TEXT
);
```

## **scenario\_lines** {#scenario_lines}

Each line is something said in the simulation.
```sql
CREATE TABLE scenario_lines (
   id UUID PRIMARY KEY,
   scenario_id UUID NOT NULL REFERENCES scenarios(id),
   sequence_no INT NOT NULL,
   speaker_type TEXT NOT NULL,
   speaker_name TEXT,
   line_text TEXT NOT NULL,
   line_kind TEXT NOT NULL CHECK (line_kind IN ('question', 'answer', 'argument', 'ruling', 'instruction')),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   UNIQUE (scenario_id, sequence_no)
);
```

## **objection\_types** {#objection_types}
```sql
CREATE TABLE objection\_types (  
   id UUID PRIMARY KEY,  
   code TEXT NOT NULL UNIQUE,  
   name TEXT NOT NULL,  
   description TEXT NOT NULL,  
   default\_phrase TEXT NOT NULL  
);
```

Example rows:
```bash
hearsay | Hearsay | Objection, hearsay.
relevance | Relevance | Objection, relevance.
foundation | Lack of Foundation | Objection, lack of foundation.
leading | Leading Question | Objection, leading.
speculation | Speculation | Objection, calls for speculation.
```

## **rule\_refs** {#rule_refs}
```sql
CREATE TABLE rule_refs (
   id UUID PRIMARY KEY,
   jurisdiction TEXT NOT NULL,
   rule_code TEXT NOT NULL,
   title TEXT NOT NULL,
   summary TEXT NOT NULL,
   source_text TEXT,
   citation TEXT NOT NULL
);
```

## **objection\_opportunities** {#objection_opportunities}

This is the answer key.
```sql
CREATE TABLE objection_opportunities (
   id UUID PRIMARY KEY,
   scenario_line_id UUID NOT NULL REFERENCES scenario_lines(id),
   objection_type_id UUID NOT NULL REFERENCES objection_types(id),
   strength TEXT NOT NULL CHECK (strength IN ('weak', 'moderate', 'strong')),
   timing_window TEXT NOT NULL CHECK (timing_window IN ('before_answer', 'after_question', 'after_answer')),
   explanation TEXT NOT NULL,
   expected_phrase TEXT,
   is_primary BOOLEAN NOT NULL DEFAULT false
)
```

## **opportunity\_rule\_refs** {#opportunity_rule_refs}
```sql
CREATE TABLE opportunity_rule_refs (
   opportunity_id UUID NOT NULL REFERENCES objection_opportunities(id),
   rule_ref_id UUID NOT NULL REFERENCES rule_refs(id),
   PRIMARY KEY (opportunity_id, rule_ref_id)
);
```

## **sessions** {#sessions}
```sql
CREATE TABLE sessions (
   id UUID PRIMARY KEY,
   user_id UUID NOT NULL REFERENCES users(id),
   scenario_id UUID NOT NULL REFERENCES scenarios(id),
   status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'abandoned')),
   current_sequence_no INT NOT NULL DEFAULT 1,
   mode TEXT NOT NULL CHECK (mode IN ('spot_objection', 'respond_to_objection', 'full_simulation')),
   started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   completed_at TIMESTAMPTZ
);
```

## **session\_events** {#session_events}

Everything important goes here.
```sql
CREATE TABLE session_events (
   id UUID PRIMARY KEY,
   session_id UUID NOT NULL REFERENCES sessions(id),
   sequence_no INT NOT NULL,
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
   metadata JSONB,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## **trainee\_actions** {#trainee_actions}
```sql
CREATE TABLE trainee_actions (
   id UUID PRIMARY KEY,
   session_id UUID NOT NULL REFERENCES sessions(id),
   scenario_line_id UUID REFERENCES scenario_lines(id),
   action_type TEXT NOT NULL CHECK (action_type IN ('object', 'respond', 'pass')),
   raw_text TEXT NOT NULL,
   normalized_objection_type_id UUID REFERENCES objection_types(id),
   created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## **action\_evaluations** {#action_evaluations}
```sql
CREATE TABLE action_evaluations (
   id UUID PRIMARY KEY,
   trainee_action_id UUID NOT NULL REFERENCES trainee_actions(id),
   matched_opportunity_id UUID REFERENCES objection_opportunities(id),
   valid BOOLEAN NOT NULL,
   timely BOOLEAN NOT NULL,
   legal_accuracy_score NUMERIC(5,2) NOT NULL,
   phrasing_score NUMERIC(5,2) NOT NULL,
   strategy_score NUMERIC(5,2) NOT NULL,
   feedback TEXT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## **session\_scores** {#session_scores}
```sql
CREATE TABLE session_scores (
   id UUID PRIMARY KEY,
   session_id UUID NOT NULL REFERENCES sessions(id),
   spotting_accuracy NUMERIC(5,2) NOT NULL,
   legal_accuracy NUMERIC(5,2) NOT NULL,
   timeliness NUMERIC(5,2) NOT NULL,
   phrasing NUMERIC(5,2) NOT NULL,
   response_quality NUMERIC(5,2) NOT NULL,
   overall_score NUMERIC(5,2) NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

# **5\. Sample event flow** {#5.-sample-event-flow}

## **Flow: trainee spots a hearsay objection** {#flow:-trainee-spots-a-hearsay-objection}

1. User starts scenario
2. Session Orchestrator creates session
3. Scenario Service returns first line
4. Frontend displays opposing counsel question
5. Frontend displays witness answer
6. Trainee clicks "Object" and types: "Objection, hearsay."
7. Session Orchestrator receives action
8. Objection Rules Service evaluates action
9. Scoring Service records result
10. LLM Simulation Service generates judge ruling language
11. Frontend displays: "Sustained."
12. Feedback Service generates short coaching note
13. Session advances to next line

## **Example transcript** {#example-transcript}
```basic
Opposing Counsel:
What did your neighbor tell you about the defendant?

Witness:
She told me the defendant admitted he caused the accident.

Trainee:
Objection, hearsay.

Judge:
Sustained.

Coach:
Correct. The witness repeated an out-of-court statement offered to prove that the defendant caused the accident. A stronger phrasing would be: "Objection, hearsay. The witness is repeating an out-of-court statement offered for its truth."
```

# **6\. API endpoints** {#6.-api-endpoints}

## **Auth** {#auth}
```typescript
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/me
```

## **Scenarios** {#scenarios-1}
```typescript
GET    /api/v1/scenarios
GET    /api/v1/scenarios/:id
POST   /api/v1/scenarios
PUT    /api/v1/scenarios/:id
POST   /api/v1/scenarios/:id/publish
```

## **Sessions** {#sessions-1}
```typescript
POST /api/v1/sessions
GET  /api/v1/sessions/:id
POST /api/v1/sessions/:id/next
POST /api/v1/sessions/:id/actions
GET  /api/v1/sessions/:id/debrief
```

## **Example create session** {#example-create-session}

`POST /api/v1/sessions`
```json
{
 "scenario_line_id": "line_004",
 "action_type": "object",
 "raw_text": "Objection, hearsay."
}
```

## **Example trainee action** {#example-trainee-action}

`POST /api/v1/sessions/session_001/actions`
```json
{
 "scenario_line_id": "line_004",
 "action_type": "object",
 "raw_text": "Objection, hearsay."
}
```

## **Example response** {#example-response}
```json
{
 "result": {
   "valid": true,
   "timely": true,
   "ruling": "sustained",
   "scores": {
     "legal_accuracy": 95,
     "phrasing": 85,
     "strategy": 90
   },
   "feedback": "Correct objection. The statement was an out-of-court assertion offered for its truth."
 },
 "next_event": {
   "speaker": "Judge",
   "text": "Sustained."
 }
}
```

# **7\. Internal service flow** {#7.-internal-service-flow}
```bash'POST /sessions/:id/actions
       |
       v
Session Orchestrator
       |
       +--> Scenario Service
       |       - get current line
       |       - get expected opportunities
       |
       +--> Objection Rules Service
       |       - classify trainee input
       |       - match to opportunity
       |       - evaluate timeliness
       |
       +--> Scoring Service
       |       - score action
       |       - update aggregates
       |
       +--> LLM Simulation Service
       |       - generate judge ruling text
       |       - generate brief coaching text
       |
       v
Return result to frontend
```

# **8\. MVP rule evaluation logic** {#8.-mvp-rule-evaluation-logic}

For the MVP, keep it simple.
```bash
Input:
- current scenario line
- trainee raw text
- expected objection opportunities
- current timing state

Steps:
1. Normalize trainee text.
2. Classify objection type.
3. Compare classified type against expected opportunity.
4. Check whether the timing window is still open.
5. Score:
  - exact match: high score
  - related objection: partial score
  - wrong objection: low score
  - valid but late: partial score
  - no valid opportunity: false positive
6. Return structured evaluation.
```

Example mapping:
```bash
"Objection, hearsay" -> hearsay
"Calls for speculation" -> speculation
"Lacks foundation" -> foundation
"Improper character evidence" -> character_evidence
```

# **9\. Suggested MVP folder structure** {#9.-suggested-mvp-folder-structure}
```bash
legal-trainer/
 backend/
   cmd/
     api/
       main.go
   internal/
     auth/
     scenarios/
     sessions/
     objections/
     scoring/
     feedback/
     llm/
     db/
     httpx/
   migrations/
   go.mod

 frontend/
   src/
     app/
     pages/
     features/
       scenarios/
       sessions/
       debrief/
     api/
     components/
   package.json

 content/
   scenarios/
     federal-rules/
       beginner-hearsay.json
       beginner-leading.json

 docs/
   architecture.md
   objection-taxonomy.md
```

# **10\. MVP build order** {#10.-mvp-build-order}

- Build in this order:
- Scenario schema and seed data
- Session creation and transcript playback
- Trainee objection submission
- Deterministic objection matching
- Judge ruling response
- Basic score calculation
- Debrief screen
- Admin scenario authoring
- LLM-enhanced coaching
- Voice mode later

The key MVP decision: **make the answer key structured first, then let AI enrich the experience.**

