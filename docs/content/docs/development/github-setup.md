+++
title = "GitHub Setup"
description = "Repository setup, branch protection, Pages"
weight = 430
icon = "settings"
toc = true
+++

# GitHub Repository Setup Guide

This guide walks you through setting up your GitHub repository with best practices, including branch protection, GitHub Pages, and automated Docker image publishing.

## Initial Repository Setup

### 1. Create Repository on GitHub

```bash
# If you haven't already, create a new repository on GitHub
# Then push your local code:

git remote add origin https://github.com/obermarclp/the-nom-database.git
git branch -M main
git push -u origin main
```

### 2. Create Develop Branch

```bash
# Create and push develop branch
git checkout -b develop
git push -u origin develop

# Set develop as default branch in GitHub Settings (optional but recommended)
```

## Branch Protection Rules

### Setting Up Branch Protection

1. Go to **Settings** → **Branches** → **Add branch protection rule**

### Protection for `main` Branch

**Branch name pattern:** `main`

**Settings:**
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: **1**
  - ✅ Dismiss stale pull request approvals when new commits are pushed
  - ✅ Require review from Code Owners (if you have CODEOWNERS file)

- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - **Required checks:**
    - `test-backend`
    - `test-frontend`
    - `security-scan`
    - `build-and-push / build-and-push (backend)`
    - `build-and-push / build-and-push (frontend)`

- ✅ **Require conversation resolution before merging**

- ✅ **Require signed commits** (optional but recommended)

- ✅ **Require linear history** (optional)

- ✅ **Do not allow bypassing the above settings**

- ❌ **Allow force pushes** (disabled)

- ❌ **Allow deletions** (disabled)

### Protection for `develop` Branch

**Branch name pattern:** `develop`

**Settings:**
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: **1**

- ✅ **Require status checks to pass before merging**
  - **Required checks:**
    - `test-backend`
    - `test-frontend`

- ✅ **Require conversation resolution before merging**

- ⚠️ **Allow force pushes**
  - Only for administrators (for occasional rebasing)

- ❌ **Allow deletions** (disabled)

## GitHub Pages Setup

### Enable GitHub Pages

1. Go to **Settings** → **Pages**

2. **Source:**
   - Select: **GitHub Actions**

3. The workflow `.github/workflows/pages.yml` will automatically deploy docs

4. **Custom domain** (optional):
   - Add your custom domain
   - Configure DNS with CNAME record

5. **Enforce HTTPS**: ✅ Enabled

### Verify Deployment

After pushing to main:
- Check **Actions** tab for "Deploy GitHub Pages" workflow
- Visit: `https://obermarclp.github.io/the-nom-database`

## GitHub Container Registry Setup

### Enable Packages

GitHub Packages is automatically enabled. Docker images will be published to:
- `ghcr.io/obermarclp/the-nom-database/backend`
- `ghcr.io/obermarclp/the-nom-database/frontend`

### Make Images Public (Recommended)

1. Go to **Packages** (from your profile or repository)
2. Click on `backend` package
3. **Package settings** → **Change visibility** → **Public**
4. Repeat for `frontend` package

### Pull Images Without Authentication

Once public, anyone can pull:

```bash
docker pull ghcr.io/obermarclp/the-nom-database/backend:latest
docker pull ghcr.io/obermarclp/the-nom-database/frontend:latest
```

## Repository Settings

### General Settings

1. Go to **Settings** → **General**

2. **Features:**
   - ✅ Wikis (if you want a wiki)
   - ✅ Issues
   - ✅ Sponsorships (optional)
   - ✅ Preserve this repository (for archiving)
   - ✅ Discussions (recommended for community)

3. **Pull Requests:**
   - ✅ Allow squash merging
   - ✅ Allow merge commits
   - ✅ Allow rebase merging
   - ✅ Always suggest updating pull request branches
   - ✅ Allow auto-merge
   - ✅ Automatically delete head branches

4. **Merge button:**
   - Default to: **Squash and merge**

### Security Settings

1. Go to **Settings** → **Code security and analysis**

2. **Dependency graph:** ✅ Enabled

3. **Dependabot:**
   - ✅ Dependabot alerts
   - ✅ Dependabot security updates

4. **Code scanning:**
   - ✅ CodeQL analysis (optional)

5. **Secret scanning:**
   - ✅ Enabled (automatically for public repos)

### Actions Settings

1. Go to **Settings** → **Actions** → **General**

2. **Actions permissions:**
   - ✅ Allow all actions and reusable workflows

3. **Workflow permissions:**
   - ✅ Read and write permissions
   - ✅ Allow GitHub Actions to create and approve pull requests

4. **Fork pull request workflows:**
   - ⚠️ Require approval for first-time contributors

## Environment Secrets

### Add Repository Secrets

Go to **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

**Optional secrets for CI/CD:**

- `CODECOV_TOKEN` - For code coverage reporting
- `VITE_API_URL` - Production API URL for frontend build

**Note:** `GITHUB_TOKEN` is automatically provided by GitHub Actions.

## Repository Labels

### Create Custom Labels

Go to **Issues** → **Labels** → **New label**

**Recommended labels:**

**Type:**
- `type: feature` - New features (color: #0075ca)
- `type: bug` - Bug fixes (color: #d73a4a)
- `type: docs` - Documentation (color: #0075ca)
- `type: refactor` - Code refactoring (color: #fbca04)
- `type: test` - Testing (color: #0e8a16)

**Priority:**
- `priority: high` - High priority (color: #d93f0b)
- `priority: medium` - Medium priority (color: #fbca04)
- `priority: low` - Low priority (color: #0e8a16)

**Status:**
- `status: needs review` - Needs review (color: #fbca04)
- `status: in progress` - In progress (color: #0075ca)
- `status: blocked` - Blocked (color: #d93f0b)

**Other:**
- `good first issue` - Good for newcomers (color: #7057ff)
- `help wanted` - Need help (color: #008672)
- `breaking change` - Breaking changes (color: #d93f0b)
- `security` - Security related (color: #d93f0b)

## Issue Templates

Create `.github/ISSUE_TEMPLATE/` directory with templates:

### Bug Report Template

Create `.github/ISSUE_TEMPLATE/bug_report.md`:

```markdown
---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: 'type: bug'
assignees: ''
---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment:**
 - OS: [e.g. iOS]
 - Browser [e.g. chrome, safari]
 - Version [e.g. 22]

**Additional context**
Add any other context about the problem here.
```

### Feature Request Template

Create `.github/ISSUE_TEMPLATE/feature_request.md`:

```markdown
---
name: Feature Request
about: Suggest an idea for this project
title: '[FEATURE] '
labels: 'type: feature'
assignees: ''
---

**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is. Ex. I'm always frustrated when [...]

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
```

## Pull Request Template

Create `.github/PULL_REQUEST_TEMPLATE.md`:

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Related Issues
Fixes #(issue number)

## How Has This Been Tested?
Please describe the tests you ran.

- [ ] Test A
- [ ] Test B

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published
```

## Release Management

### Creating Your First Release

```bash
# Make sure you're on main branch
git checkout main
git pull origin main

# Create a tag
git tag -a v1.0.0 -m "Initial release v1.0.0"

# Push tag (this triggers release workflow)
git push origin v1.0.0
```

### GitHub Actions will automatically:
1. ✅ Build Docker images for linux/amd64 and linux/arm64
2. ✅ Publish images to GitHub Container Registry
3. ✅ Create GitHub Release with changelog
4. ✅ Build platform-specific binaries
5. ✅ Upload binaries to release

### Release Page

Visit: `https://github.com/obermarclp/the-nom-database/releases`

## Webhooks (Optional)

### Discord Notifications

1. Create webhook in Discord server
2. Go to **Settings** → **Webhooks** → **Add webhook**
3. **Payload URL**: Your Discord webhook URL
4. **Content type**: application/json
5. **Events**: Select what to notify

### Slack Notifications

Similar to Discord, integrate with Slack workspace.

## Repository Insights

### Enable Insights

1. Go to **Insights** tab
2. Useful views:
   - **Pulse** - Activity overview
   - **Contributors** - Contribution stats
   - **Traffic** - Views and clones
   - **Dependency graph** - Dependencies
   - **Network** - Forks and branches

## Community Health Files

Create these files in repository root:

### CODE_OF_CONDUCT.md

Define community standards and behavior.

### SECURITY.md

```markdown
# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

Please report security vulnerabilities to security@yourdomain.com or via GitHub Security Advisories.

**Do not** open public issues for security vulnerabilities.
```

### CODEOWNERS (Optional)

Create `.github/CODEOWNERS`:

```
# Default owners
* @obermarclp

# Backend code
/backend/ @backend-team

# Frontend code
/frontend/ @frontend-team

# Documentation
/docs/ @docs-team

# CI/CD
/.github/ @devops-team
```

## Verification Checklist

After setup, verify:

- [ ] Repository created and code pushed
- [ ] `main` and `develop` branches exist
- [ ] Branch protection rules configured
- [ ] GitHub Pages deployed
- [ ] Actions tab shows successful workflows
- [ ] Docker images published to ghcr.io
- [ ] Images are public (if intended)
- [ ] Issue/PR templates working
- [ ] Labels created
- [ ] First release tagged
- [ ] Documentation accessible

## Maintenance

### Regular Tasks

**Weekly:**
- Review and merge Dependabot PRs
- Triage new issues
- Review open PRs

**Monthly:**
- Check GitHub Actions usage/costs
- Review security alerts
- Update documentation

**Quarterly:**
- Audit branch protection rules
- Review and clean up stale branches
- Analyze repository insights

## Troubleshooting

### GitHub Actions Failing

1. Check Actions tab for error logs
2. Verify secrets are set correctly
3. Check branch protection status checks

### Docker Images Not Publishing

1. Verify workflow permissions (Settings → Actions)
2. Check if images are set to public
3. Verify GITHUB_TOKEN permissions

### GitHub Pages Not Deploying

1. Check Pages workflow in Actions
2. Verify Pages is enabled in Settings
3. Check for Jekyll build errors

## Additional Resources

- [GitHub Docs](https://docs.github.com)
- [GitHub Actions Marketplace](https://github.com/marketplace?type=actions)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Support

For repository setup questions:
- Open a Discussion
- Contact repository maintainers
- Check GitHub Documentation

---

**Next Steps:** Once setup is complete, see [Git Workflow](/docs/development/git-workflow/) for daily development workflows.
