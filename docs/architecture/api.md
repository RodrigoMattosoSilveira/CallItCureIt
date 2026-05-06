# Introduction
We will use OpenAPI for this project, especially because our MVP has clear service boundaries, typed API contracts, and a frontend/backend split.

For this project, I would use OpenAPI 3 as the contract between the React frontend and the backend API, not as the primary model for our internal legal reasoning logic.
# Why OpenAPI 3 fits well
This project has several API-heavy areas:
```bash
Scenarios
Sessions
Trainee actions
Judge rulings
Debriefs
Scores
Admin content management
```

Those are perfect for OpenAPI because they involve structured request/response contracts.
For example:

```typescript
POST /api/v1/sessions
POST /api/v1/sessions/{sessionId}/actions
GET  /api/v1/sessions/{sessionId}/debrief
GET  /api/v1/scenarios
```

OpenAPI lets us define these once and then generate or validate:
- TypeScript frontend clients
- backend request/response models
- API documentation
- contract tests
- mock servers
- SDKs later
- validation middleware
# Where OpenAPI helps most
## 1. Frontend/backend contract
Our React app will need strongly typed calls like:
```typescript
createSession(input)
submitTraineeAction(sessionId, input)
getDebrief(sessionId)
listScenarios()
```

With OpenAPI, us can generate those clients instead of manually duplicating DTOs in TypeScript and Go.
That prevents the common problem:
```bash
Backend returns `scenario_id`
Frontend expects `scenarioId`
```

or:
```bash
Backend changes response shape
Frontend silently breaks
```

## 2. Better domain modeling
Our API has domain objects that should be explicit:
- Scenario
- ScenarioLine
- ObjectionType
- ObjectionOpportunity
- Session
- SessionEvent
- TraineeAction
- ActionEvaluation
- JudgeRuling
- Debrief
- SessionScore

OpenAPI forces us to define those clearly. Example:
```yaml
TraineeActionInput:
 type: object
 required:
   - scenarioLineId
   - actionType
   - rawText
 properties:
   scenarioLineId:
     type: string
     format: uuid
   actionType:
     type: string
     enum:
       - object
       - respond
       - pass
   rawText:
     type: string
```
That is very useful for a legal training system because vague payloads will become painful quickly.

## 3. Easier testing
We wil use OpenAPI for:
- request validation
- response validation
- contract testing
- mock APIs
- frontend development before backend endpoints are complete

For our style of development, this is especially helpful because we often like production-grade examples with tests and clear contracts.
## 4. Cleaner admin tooling
The admin/instructor side will eventually need endpoints for:
- Create scenario
- Edit scenario lines
- Add expected objections
- Attach rule references
- Publish scenario
- Archive scenario

OpenAPI will keep those endpoints understandable as the system grows.

# Where OpenAPI should not be overused
Do not try to model the internal reasoning engine entirely in OpenAPI.

For example, these internals should probably remain service/domain logic:
- Does this line contain hearsay?
- Was the objection timely?
- Was the trainee response strategically strong?
- Should the judge sustain or overrule?

OpenAPI should describe the external API boundary, not every internal decision branch.

# Recommended API style
We will use a mostly RESTful OpenAPI 3 design. Good MVP endpoints:
```typescript
GET    /api/v1/scenarios
GET    /api/v1/scenarios/{scenarioId}

POST   /api/v1/sessions
GET    /api/v1/sessions/{sessionId}
POST   /api/v1/sessions/{sessionId}/next
POST   /api/v1/sessions/{sessionId}/actions
POST   /api/v1/sessions/{sessionId}/complete

GET    /api/v1/sessions/{sessionId}/debrief
GET    /api/v1/sessions/{sessionId}/score

GET    /api/v1/objection-types
GET    /api/v1/rule-refs

// Admin endpoints:
POST   /api/v1/admin/scenarios
PUT    /api/v1/admin/scenarios/{scenarioId}
POST   /api/v1/admin/scenarios/{scenarioId}/publish
POST   /api/v1/admin/scenarios/{scenarioId}/archive

POST   /api/v1/admin/scenarios/{scenarioId}/lines
PUT    /api/v1/admin/scenario-lines/{lineId}

POST   /api/v1/admin/scenario-lines/{lineId}/objection-opportunities
PUT    /api/v1/admin/objection-opportunities/{opportunityId}
```

# Recommended tooling for a Go + React project
Given our usual stack, I would consider:
## Backend
For Go, a good option is:
`oapi-codegen`

Use it to generate:
- Go server interfaces
- request/response types
- validation helpers

Suggested structure:
```bash
backend/
 api/
   openapi.yaml
   generated/
     api.gen.go
 internal/
   sessions/
   scenarios/
   objections/
   scoring/
```

## Frontend
For React + TypeScript, generate a typed client from the same spec.
Options:
- openapi-typescript
- openapi-fetch
- orval

We will use the combination:
`openapi-typescript + openapi-fetch`

This gives us strong TypeScript types without generating an overly heavy client.

# OpenAPI-first or code-first?
For this project, we will take an OpenAPI-first approach. That means:
- define openapi.yaml
- generate backend interfaces/types
- generate frontend client/types
- implement backend handlers
- write contract tests

Why OpenAPI-first is better here:
- our frontend and backend are separate
- the domain has many structured DTOs
- legal training flows need stable contracts
- we may eventually have mobile, instructor dashboards, or third-party integrations

# Suggested MVP OpenAPI modules
Do not put everything into one massive file long-term.
Start with:
```bash
api/
 openapi.yaml
```

Later split into:
```bash
api/
 openapi.yaml
 paths/
   scenarios.yaml
   sessions.yaml
   admin.yaml
 schemas/
   scenario.yaml
   session.yaml
   objection.yaml
   scoring.yaml
   debrief.yaml
```

# Example schema
```yaml
ActionEvaluation:
 type: object
 required:
   - valid
   - timely
   - ruling
   - scores
   - feedback
 properties:
   valid:
     type: boolean
   timely:
     type: boolean
   ruling:
     type: string
     enum:
       - sustained
       - overruled
       - partially_sustained
       - no_ruling
   matchedOpportunityId:
     type: string
     format: uuid
     nullable: true
   scores:
     $ref: '#/components/schemas/ActionScores'
   feedback:
     type: string

ActionScores:
 type: object
 required:
   - legalAccuracy
   - phrasing
   - strategy
 properties:
   legalAccuracy:
     type: number
     minimum: 0
     maximum: 100
   phrasing:
     type: number
     minimum: 0
     maximum: 100
   strategy:
     type: number
     minimum: 0
     maximum: 100
```

# Important recommendation
Use OpenAPI for public HTTP contracts, but keep our core legal evaluation model separate as internal domain types.
For example:
```texdt
OpenAPI DTO:
SubmitTraineeActionRequest

Internal domain command:
EvaluateObjectionCommand

Internal domain result:
ObjectionEvaluationResult
```

That gives us flexibility. Our API can remain stable even if our internal scoring engine evolves.

# Recommendation
Use OpenAPI 3.1, OpenAPI-first, with generated Go and TypeScript types.
For the MVP, define contracts for:
- Scenarios
- Sessions
- Actions
- Evaluations
- Scores
- Debriefs
- - Objection types
Rule references

Then keep the objection reasoning engine behind those contracts.

That gives us a clean, testable, frontend-friendly architecture without locking our internal legal logic into an API specification.
