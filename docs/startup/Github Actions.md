# Introduction
This is written around our current working server model: GitHub merges trigger branch-specific deployment commands over SSH, while each environment keeps its own server checkout and Makefile-driven launch sequence.

Below is a clean GitHub Actions deployment pipeline for your current Call It Cure It architecture.

It deploys automatically when changes are merged/pushed into:

```bash
development -> dev.callitcureit.com
test        -> tst.callitcureit.com
production  -> app.callitcureit.com
```

```bash
**It assumes your server already has**:

/opt/CallItCureIt/
  development/
  test/
  production/
  edge/
```

and each environment folder has its own clone of the repo.

# 1. GitHub Actions Secrets

Create these repository secrets:

```
DEPLOY_HOST
DEPLOY_USER
DEPLOY_SSH_PRIVATE_KEY
DEPLOY_KNOWN_HOSTS
```

**Example values:**

```env
DEPLOY_HOST=5.78.208.230
DEPLOY_USER=deploy
```

`DEPLOY_SSH_PRIVATE_KEY` is the private key GitHub Actions uses to SSH into the Hetzner server.

`DEPLOY_KNOWN_HOSTS` should contain the verified SSH host key for the server. Generate it from your local machine or the server console:

ssh-keyscan -H 5.78.208.230

Then paste the output into the GitHub secret.

# 2. Server-side prerequisites

The GitHub Action will SSH into the server and run git pull / make.

So each server folder must already exist:

```bash
/opt/CallItCureIt/development
/opt/CallItCureIt/test
/opt/CallItCureIt/production
```

**Each one should be checked out to the matching branch:**

```bash
cd /opt/CallItCureIt/development
git checkout development

cd /opt/CallItCureIt/test
git checkout test

cd /opt/CallItCureIt/production
git checkout production
```

The server itself must be able to pull from GitHub. That means your server deploy key must already be configured.

# 3. GitHub Actions Workflow

**Create this file:**

```yaml
name: Deploy Call It Cure It

run-name: Deploy ${{ github.event_name == 'workflow_dispatch' && inputs.environment || github.ref_name }}

on:
  push:
    branches:
      - development
      - test
      - production

  workflow_dispatch:
    inputs:
      environment:
        description: "Environment to deploy"
        required: true
        type: choice
        options:
          - development
          - test
          - production
      deploy_edge:
        description: "Also sync/reload the edge proxy from this branch"
        required: false
        default: false
        type: boolean

permissions:
  contents: read

concurrency:
  group: deploy-${{ github.event_name == 'workflow_dispatch' && inputs.environment || github.ref_name }}
  cancel-in-progress: false

jobs:
  deploy:
    name: Deploy environment
    runs-on: ubuntu-latest
    timeout-minutes: 45

    environment: ${{ github.event_name == 'workflow_dispatch' && inputs.environment || github.ref_name }}

    env:
      SERVER_ROOT: /opt/CallItCureIt
      DEPLOY_HOST: ${{ secrets.DEPLOY_HOST }}
      DEPLOY_USER: ${{ secrets.DEPLOY_USER }}

    steps:
      - name: Resolve deployment target
        id: target
        shell: bash
        run: |
          set -euo pipefail

          if [[ "${{ github.event_name }}" == "workflow_dispatch" ]]; then
            TARGET_ENV="${{ inputs.environment }}"
            MANUAL_EDGE="${{ inputs.deploy_edge }}"
          else
            TARGET_ENV="${GITHUB_REF_NAME}"
            MANUAL_EDGE="false"
          fi

          case "$TARGET_ENV" in
            development)
              echo "env_name=development" >> "$GITHUB_OUTPUT"
              echo "branch=development" >> "$GITHUB_OUTPUT"
              echo "env_dir=${SERVER_ROOT}/development" >> "$GITHUB_OUTPUT"
              echo "make_prefix=server-dev" >> "$GITHUB_OUTPUT"
              echo "domain=dev.callitcureit.com" >> "$GITHUB_OUTPUT"
              echo "deploy_edge=${MANUAL_EDGE}" >> "$GITHUB_OUTPUT"
              ;;
            test)
              echo "env_name=test" >> "$GITHUB_OUTPUT"
              echo "branch=test" >> "$GITHUB_OUTPUT"
              echo "env_dir=${SERVER_ROOT}/test" >> "$GITHUB_OUTPUT"
              echo "make_prefix=server-test" >> "$GITHUB_OUTPUT"
              echo "domain=tst.callitcureit.com" >> "$GITHUB_OUTPUT"
              echo "deploy_edge=${MANUAL_EDGE}" >> "$GITHUB_OUTPUT"
              ;;
            production)
              echo "env_name=production" >> "$GITHUB_OUTPUT"
              echo "branch=production" >> "$GITHUB_OUTPUT"
              echo "env_dir=${SERVER_ROOT}/production" >> "$GITHUB_OUTPUT"
              echo "make_prefix=server-prod" >> "$GITHUB_OUTPUT"
              echo "domain=app.callitcureit.com" >> "$GITHUB_OUTPUT"

              # Production is the canonical branch for shared edge proxy deployment.
              # Manual dispatch can also explicitly request deploy_edge=true.
              echo "deploy_edge=true" >> "$GITHUB_OUTPUT"
              ;;
            *)
              echo "Unsupported target environment: $TARGET_ENV"
              exit 1
              ;;
          esac

      - name: Configure SSH
        shell: bash
        run: |
          set -euo pipefail

          mkdir -p ~/.ssh
          chmod 700 ~/.ssh

          cat > ~/.ssh/id_ed25519 <<'EOF'
          ${{ secrets.DEPLOY_SSH_PRIVATE_KEY }}
          EOF

          chmod 600 ~/.ssh/id_ed25519

          cat > ~/.ssh/known_hosts <<'EOF'
          ${{ secrets.DEPLOY_KNOWN_HOSTS }}
          EOF

          chmod 644 ~/.ssh/known_hosts

      - name: Preflight server checks
        shell: bash
        run: |
          set -euo pipefail

          ssh -i ~/.ssh/id_ed25519 \
            "${DEPLOY_USER}@${DEPLOY_HOST}" \
            "set -euo pipefail
             test -d '${{ steps.target.outputs.env_dir }}'
             test -f '${{ steps.target.outputs.env_dir }}/Makefile'
             test -f '${{ steps.target.outputs.env_dir }}/docker-compose.server.yml'
             test -f '${{ steps.target.outputs.env_dir }}/deploy/Caddyfile'
             echo 'Preflight checks passed for ${{ steps.target.outputs.env_name }}'
            "

      - name: Deploy app stack
        shell: bash
        run: |
          set -euo pipefail

          ssh -i ~/.ssh/id_ed25519 \
            "${DEPLOY_USER}@${DEPLOY_HOST}" \
            "set -euo pipefail

             cd '${{ steps.target.outputs.env_dir }}'

             echo 'Deploying ${{ steps.target.outputs.env_name }} from branch ${{ steps.target.outputs.branch }}'
             git fetch origin '${{ steps.target.outputs.branch }}'
             git checkout '${{ steps.target.outputs.branch }}'
             git reset --hard 'origin/${{ steps.target.outputs.branch }}'

             make ${{ steps.target.outputs.make_prefix }}-build
             make ${{ steps.target.outputs.make_prefix }}-up
             make ${{ steps.target.outputs.make_prefix }}-backend-health
            "

      - name: Deploy edge proxy when required
        if: ${{ steps.target.outputs.deploy_edge == 'true' }}
        shell: bash
        run: |
          set -euo pipefail

          ssh -i ~/.ssh/id_ed25519 \
            "${DEPLOY_USER}@${DEPLOY_HOST}" \
            "set -euo pipefail

             cd '${{ steps.target.outputs.env_dir }}'

             echo 'Deploying edge proxy from ${{ steps.target.outputs.branch }} branch'
             make edge-sync
             make edge-up
             make edge-reload
            "

      - name: Public smoke tests
        shell: bash
        run: |
          set -euo pipefail

          ssh -i ~/.ssh/id_ed25519 \
            "${DEPLOY_USER}@${DEPLOY_HOST}" \
            "set -euo pipefail

             cd '${{ steps.target.outputs.env_dir }}'

             echo 'Running smoke tests for ${{ steps.target.outputs.domain }}'
             make ${{ steps.target.outputs.make_prefix }}-smoke
             make ${{ steps.target.outputs.make_prefix }}-admin-test
            "

      - name: Deployment summary
        if: always()
        shell: bash
        run: |
          {
            echo "## Deployment Summary"
            echo
            echo "| Item | Value |"
            echo "|---|---|"
            echo "| Environment | ${{ steps.target.outputs.env_name }} |"
            echo "| Branch | ${{ steps.target.outputs.branch }} |"
            echo "| Domain | ${{ steps.target.outputs.domain }} |"
            echo "| Server folder | ${{ steps.target.outputs.env_dir }} |"
            echo "| Make prefix | ${{ steps.target.outputs.make_prefix }} |"
            echo "| Edge deployed | ${{ steps.target.outputs.deploy_edge }} |"
          } >> "$GITHUB_STEP_SUMMARY"
```

# 4. Important Edge Proxy Policy

Because your edge proxy is shared by all three environments, I recommend this policy:

**Normal app deployments:**
- development branch deploys only the development app stack.
- test branch deploys only the test app stack.
- production branch deploys the production app stack and syncs/reloads edge.

Why?

Because the edge proxy affects all three public domains:

```
dev.callitcureit.com
tst.callitcureit.com
app.callitcureit.com
```

So it is safer for automatic edge deployment to happen from the production branch only.

**If you intentionally want to test an edge change from development or test, use the manual workflow and set:
**
`deploy_edge=true`

# 5. Recommended Branch Protection

**In GitHub, protect these branches:**
```
development
test
production
```

**Minimum recommendation:**

`development`:
- require pull request before merging
- require status checks if you add tests later

`test`:
- require pull request before merging
- require development validation before PR

`production`:
- require pull request before merging
- require test validation before PR
- optionally require manual approval through GitHub Environments

This workflow triggers on push to those branches. In practice, with branch protection, that means it triggers after a PR merge.

# 6. Optional: Add a CI Test Workflow Before Deploy
`.github/workflows/ci.yml`
```yaml
name: CI

on:
  pull_request:
    branches:
      - development
      - test
      - production

permissions:
  contents: read

jobs:
  backend:
    name: Backend tests
    runs-on: ubuntu-latest

    defaults:
      run:
        working-directory: backend

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.26"

      - name: Test backend
        run: go test ./...

  frontend:
    name: Frontend checks
    runs-on: ubuntu-latest

    defaults:
      run:
        working-directory: frontend

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: "22"
          cache: npm
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: Run checks
        run: npm run check
```

# 7. Exact Deployment Behavior
## Merge into development

**GitHub Actions runs:**
`target: https://dev.callitcureit.com`
```bash
cd /opt/CallItCureIt/development
git fetch origin development
git checkout development
git reset --hard origin/development
make server-dev-build
make server-dev-up
make server-dev-backend-health
make server-dev-smoke
make server-dev-admin-test
```

## Merge into test

**GitHub Actions runs:**
`target: https://tst.callitcureit.com`
```bash
cd /opt/CallItCureIt/test
git fetch origin test
git checkout test
git reset --hard origin/test
make server-tst-build
make server-tst-up
make server-tst-backend-health
make server-tst-smoke
make server-tst-admin-test
```

## Merge into production

**GitHub Actions runs:**
`target: https://app.callitcureit.com`
```bash
cd /opt/CallItCureIt/production
git fetch origin production
git checkout production
git reset --hard origin/production
make server-prod-build
make server-prod-up
make server-prod-backend-health
make edge-sync
make edge-up
make edge-reload
make server-prod-smoke
```
# 8. First Manual Run

After committing the workflow, you can manually run it from GitHub:
```
Actions
  -> Deploy Call It Cure It
  -> Run workflow
  -> environment: development
  -> deploy_edge: false
```
Then:
```
environment: test
deploy_edge: false
```
Then:
```
environment: production
deploy_edge: true
```
Once that works, merges into the three branches will deploy automatically.