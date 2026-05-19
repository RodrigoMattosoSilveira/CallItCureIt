# Introduction
**Pull requests into development/test/production:**
  - Go tests
  - Frontend checks
  - Local Playwright E2E against disposable CI app

**Merges/pushes into development:**
  - Quality gate
  - Deploy dev
  - Smoke/admin checks
  - Playwright against https://dev.callitcureit.com

**Merges/pushes into test:**
  - Quality gate
  - Deploy test
  - Smoke/admin checks
  - Playwright against https://tst.callitcureit.com

**Merges/pushes into production:**
  - Quality gate
  - Deploy production
  - Edge deploy/reload
  - Smoke/admin-read checks only
  - No deployed Playwright

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
Note the `pre-deploy` quality gate to precent a bad direct push or mistaken merge to deploy.
`.github/workflows/ci.yml`

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
Note the 
`.github/workflows/ci.yml`

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

# Set up the Hetzner server
We need two different SSH trust pieces:

`DEPLOY_SSH_PRIVATE_KEY`
```
The private key GitHub Actions uses to SSH into your Hetzner server.
```

`DEPLOY_KNOWN_HOSTS`
```
The server host key fingerprint GitHub Actions uses to verify it is connecting to the real Hetzner server.
```

GitHub recommends storing sensitive values as `repository` or `environment` secrets, and we need repository write/admin access to create repository secrets.

```
We will store the DEPLOY_SSH_PRIVATE_KEY the DEPLOY_KNOWN_HOSTS as repository secrets.
```

## 1. Generate a dedicated GitHub Actions deploy key

Do this on your local machine, not inside the repo.

```bash
mkdir -p ~/.ssh/callitcureit-github-actions
cd ~/.ssh/callitcureit-github-actions

ssh-keygen -t ed25519 \
  -C "github-actions-callitcureit-deploy" \
  -f callitcureit_github_actions_deploy
```

When prompted for a passphrase, press Enter for no passphrase.

You should now have:

```
callitcureit_github_actions_deploy      private key
callitcureit_github_actions_deploy.pub  public key
```

Do not commit either file.

## 2. Install the public key on the Hetzner server

**Copy the public key to the server:**
```bash
ssh-copy-id \
  -i ~/.ssh/callitcureit-github-actions/callitcureit_github_actions_deploy.pub \
  deploy@5.78.208.230
```

**If ssh-copy-id is not available, use this manual method:**
```bash
cat ~/.ssh/callitcureit-github-actions/callitcureit_github_actions_deploy.pub
```
Copy the output.

Then SSH into the server as deploy and append it:
```bash
ssh deploy@5.78.208.230

mkdir -p ~/.ssh
chmod 700 ~/.ssh

nano ~/.ssh/authorized_keys

# Paste the public key as one line at the bottom, save, then run:

chmod 600 ~/.ssh/authorized_keys

# Exit:

exit
```

## 3. Test the key from your local machine

**Run:**

```bash
ssh -i ~/.ssh/callitcureit-github-actions/callitcureit_github_actions_deploy \
  deploy@5.78.208.230 \
  'hostname && whoami && docker ps --format "table {{.Names}}\t{{.Status}}"'
```
**Expected:**
```
call-it-cure-it
deploy
<docker container list>
```

If this fails, do not continue to GitHub secrets yet.

## 4. Create DEPLOY_SSH_PRIVATE_KEY

**Show the private key:**
```bash
cat ~/.ssh/callitcureit-github-actions/callitcureit_github_actions_deploy
```

**Copy the entire output, including the first and last lines:**
```
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```
**Then in GitHub:**
```
Repository
  → Settings
  → Secrets and variables
  → Actions
  → New repository secret
```
**Create:**
```
Name:
DEPLOY_SSH_PRIVATE_KEY

Secret:
<paste the entire private key>
```
**GitHub’s Actions secrets UI is the correct place for this kind of value.**

## 5. Create DEPLOY_KNOWN_HOSTS

**Run this on your local machine:**
```bash
ssh-keyscan -H 5.78.208.230
```

**You may also include the hostname if you SSH by hostname:**
```bash
ssh-keyscan -H call-it-cure-it 5.78.208.230
```

**For our workflow, since DEPLOY_HOST=5.78.208.230, this is enough:**
```bash
ssh-keyscan -H 5.78.208.230
```
**Copy the full output. It will look similar to:**
```
|1|... ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA...
|1|... ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTIt...
|1|... ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB...
```
**Then in GitHub:**
```
Repository
  → Settings
  → Secrets and variables
  → Actions
  → New repository secret
```
**Create:**
```
Name:
DEPLOY_KNOWN_HOSTS
```
Secret:
<paste the full ssh-keyscan output>
```
This lets the workflow write a safe ~/.ssh/known_hosts file instead of disabling host verification.

## 6. Create the remaining deploy secrets

**In the same GitHub Actions secrets area, create:**
`DEPLOY_HOST`

**Value:**

`5.78.208.230`

**Create:**

`DEPLOY_USER`

**Value:**

`deploy`

So your four required repository secrets are:
```
DEPLOY_HOST
DEPLOY_USER
DEPLOY_SSH_PRIVATE_KEY
DEPLOY_KNOWN_HOSTS
```

## 7. Verify the secrets with a manual workflow run

After committing deploy.yml, go to:
```
GitHub repository
  → Actions
  → Deploy Call It Cure It
  → Run workflow
```
**Choose:**
```
environment: development
deploy_edge: false
run_deployed_playwright: true
```
If SSH is configured correctly, the workflow should pass the step:

Preflight server checks

## 8. Common mistakes
### Mistake 1 — storing the public key instead of private key

DEPLOY_SSH_PRIVATE_KEY must contain:

-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----

Not the .pub file.

### Mistake 2 — adding the private key to the server

The server gets the public key in:

/home/deploy/.ssh/authorized_keys

GitHub gets the private key in:

DEPLOY_SSH_PRIVATE_KEY

### Mistake 3 — passphrase-protected key

For this workflow, use a deploy key with no passphrase, unless you also design the workflow to handle the passphrase.

### Mistake 4 — wrong known hosts value

DEPLOY_KNOWN_HOSTS must match the value used by:

DEPLOY_HOST

If DEPLOY_HOST=5.78.208.230, generate known hosts with:

ssh-keyscan -H 5.78.208.230

### Mistake 5 — server user cannot run Docker

On the server, the deploy user must be able to run Docker without sudo:

docker ps

If not:

sudo usermod -aG docker deploy

Then log out and back in.

## 9. Optional safer approach: GitHub Environments
**For extra safety, store the same secrets under GitHub Environments:**
```
development
test
production
```

This lets you add approval rules for production. 

GitHub supports repository, environment, and organization secrets for Actions. 

`For now, repository secrets are simpler and fine.`
