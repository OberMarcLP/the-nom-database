+++
title = "Git Workflow"
description = "Branching strategy and release process"
weight = 420
icon = "account_tree"
toc = true
+++

# Git Workflow and Release Process

This document describes the branching strategy, development workflow, and release process for The Nom Database.

## Branching Strategy

We follow **Git Flow** with some modifications for simplicity.

### Main Branches

#### `main`
- **Purpose**: Production-ready code
- **Protection**: Protected, requires PR reviews
- **Deployment**: Automatically deploys to production
- **Tags**: All release tags are created from main
- **Never commit directly to main**

#### `develop`
- **Purpose**: Integration branch for features
- **Protection**: Protected, requires PR reviews
- **Deployment**: Automatically deploys to staging/development
- **Merge from**: Feature branches, bugfix branches
- **Merge to**: Main (via release PR)

### Supporting Branches

#### Feature Branches
- **Naming**: `feature/description` or `feature/issue-number-description`
- **Purpose**: Develop new features
- **Branch from**: `develop`
- **Merge to**: `develop`
- **Lifetime**: Deleted after merge

**Examples:**
- `feature/add-user-profiles`
- `feature/123-restaurant-search`
- `feature/oidc-authentication`

#### Bugfix Branches
- **Naming**: `bugfix/description` or `bugfix/issue-number-description`
- **Purpose**: Fix bugs in develop branch
- **Branch from**: `develop`
- **Merge to**: `develop`
- **Lifetime**: Deleted after merge

**Examples:**
- `bugfix/fix-rating-calculation`
- `bugfix/456-photo-upload-error`

#### Hotfix Branches
- **Naming**: `hotfix/description` or `hotfix/vX.Y.Z`
- **Purpose**: Critical fixes for production
- **Branch from**: `main`
- **Merge to**: `main` AND `develop`
- **Lifetime**: Deleted after merge

**Examples:**
- `hotfix/security-patch`
- `hotfix/v1.0.1`

#### Release Branches
- **Naming**: `release/vX.Y.Z`
- **Purpose**: Prepare for production release
- **Branch from**: `develop`
- **Merge to**: `main` AND `develop`
- **Lifetime**: Deleted after merge

**Examples:**
- `release/v1.0.0`
- `release/v1.1.0`

## Workflow Examples

### 1. Developing a New Feature

```bash
# Update develop branch
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/add-user-profiles

# Make changes, commit regularly
git add .
git commit -m "feat: add user profile page"
git commit -m "feat: add profile edit functionality"
git commit -m "test: add profile page tests"

# Push to remote
git push origin feature/add-user-profiles

# Create Pull Request on GitHub
# After approval and CI passes, merge to develop
# Delete feature branch
```

### 2. Fixing a Bug

```bash
# Update develop branch
git checkout develop
git pull origin develop

# Create bugfix branch
git checkout -b bugfix/fix-rating-calculation

# Fix the bug
git add .
git commit -m "fix: correct average rating calculation"

# Push and create PR
git push origin bugfix/fix-rating-calculation

# After merge, delete branch
```

### 3. Creating a Release

```bash
# Update develop branch
git checkout develop
git pull origin develop

# Create release branch
git checkout -b release/v1.0.0

# Update version numbers
# Update CHANGELOG.md
# Final testing and bug fixes
git commit -am "chore: prepare v1.0.0 release"

# Push release branch
git push origin release/v1.0.0

# Create PR to main
# After approval, merge to main
git checkout main
git merge release/v1.0.0

# Tag the release
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Merge back to develop
git checkout develop
git merge release/v1.0.0
git push origin develop

# Delete release branch
git branch -d release/v1.0.0
git push origin --delete release/v1.0.0
```

### 4. Emergency Hotfix

```bash
# Create hotfix from main
git checkout main
git pull origin main
git checkout -b hotfix/v1.0.1

# Fix the critical issue
git commit -am "fix: patch critical security vulnerability"

# Push hotfix branch
git push origin hotfix/v1.0.1

# Merge to main
git checkout main
git merge hotfix/v1.0.1

# Tag the hotfix
git tag -a v1.0.1 -m "Hotfix version 1.0.1"
git push origin v1.0.1

# Merge to develop
git checkout develop
git merge hotfix/v1.0.1
git push origin develop

# Delete hotfix branch
git branch -d hotfix/v1.0.1
git push origin --delete hotfix/v1.0.1
```

## Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/).

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, no logic change)
- **refactor**: Code refactoring
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **build**: Build system or dependency changes
- **ci**: CI/CD changes
- **chore**: Other changes (e.g., releasing)

### Examples

```bash
# Simple feature
git commit -m "feat: add restaurant photo gallery"

# Bug fix with scope
git commit -m "fix(api): correct rating calculation logic"

# Breaking change
git commit -m "feat!: migrate to OIDC authentication

BREAKING CHANGE: Google OAuth has been replaced with generic OIDC.
Update your configuration to use OIDC_* environment variables."

# With issue reference
git commit -m "fix: resolve photo upload timeout (#123)"

# Multiple changes
git commit -m "feat: add user profiles

- Add profile page
- Add profile edit form
- Add avatar upload
- Update API endpoints"
```

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes (e.g., v1.0.0 → v2.0.0)
- **MINOR**: New features, backwards compatible (e.g., v1.0.0 → v1.1.0)
- **PATCH**: Bug fixes, backwards compatible (e.g., v1.0.0 → v1.0.1)

### Pre-release Versions

- **Alpha**: `v1.0.0-alpha.1` - Early testing
- **Beta**: `v1.0.0-beta.1` - Feature complete, testing
- **RC**: `v1.0.0-rc.1` - Release candidate

### Release Checklist

#### Before Release

- [ ] All features merged to develop
- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated
- [ ] Migration scripts tested
- [ ] Security scan passed

#### Creating Release

1. **Create release branch**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b release/vX.Y.Z
   ```

2. **Prepare release**
   ```bash
   # Update version in files
   # Update CHANGELOG.md
   # Final testing
   git commit -am "chore: prepare vX.Y.Z release"
   ```

3. **Create PR to main**
   - Review changes
   - Run all tests
   - Get approval

4. **Merge and tag**
   ```bash
   git checkout main
   git merge release/vX.Y.Z
   git tag -a vX.Y.Z -m "Release version X.Y.Z"
   git push origin main
   git push origin vX.Y.Z
   ```

5. **GitHub Actions will automatically:**
   - Build Docker images
   - Publish to GitHub Container Registry
   - Create GitHub Release
   - Generate changelog
   - Build platform binaries

6. **Merge back to develop**
   ```bash
   git checkout develop
   git merge release/vX.Y.Z
   git push origin develop
   ```

7. **Clean up**
   ```bash
   git branch -d release/vX.Y.Z
   git push origin --delete release/vX.Y.Z
   ```

#### After Release

- [ ] Verify Docker images published
- [ ] Test deployed containers
- [ ] Update documentation site
- [ ] Announce release
- [ ] Monitor for issues

## Pull Request Guidelines

### Creating a PR

1. **Title**: Use conventional commit format
   - `feat: add user authentication`
   - `fix: resolve rating calculation bug`

2. **Description**: Include
   - What changed and why
   - How to test
   - Screenshots (if UI changes)
   - Related issues

3. **Labels**: Add appropriate labels
   - `feature`, `bugfix`, `hotfix`, `documentation`
   - `breaking-change`, `security`

4. **Reviewers**: Request reviews from maintainers

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts
- [ ] CI checks passing
- [ ] Self-reviewed code

### Review Process

1. **Automated checks** (GitHub Actions)
   - Build passes
   - Tests pass
   - Linting passes
   - Security scan passes

2. **Code review** (at least 1 approval)
   - Code quality
   - Test coverage
   - Documentation
   - Security considerations

3. **Merge**
   - Squash and merge (for feature branches)
   - Merge commit (for release/hotfix to main)

## Branch Protection Rules

### Main Branch
- Require pull request reviews (1 approval minimum)
- Require status checks to pass
- Require branches to be up to date
- Do not allow force pushes
- Do not allow deletions

### Develop Branch
- Require pull request reviews (1 approval minimum)
- Require status checks to pass
- Allow force pushes for maintainers only
- Do not allow deletions

## Docker Image Tags

Docker images are automatically tagged based on git activity:

### Automatic Tags

- **`latest`**: Latest commit on main branch
- **`develop`**: Latest commit on develop branch
- **`vX.Y.Z`**: Exact version tag (e.g., v1.0.0)
- **`vX.Y`**: Latest patch version (e.g., v1.0)
- **`vX`**: Latest minor version (e.g., v1)
- **`main-abc123`**: Commit SHA from main
- **`pr-123`**: Pull request number

### Using Tags

```bash
# Production (use specific version)
docker pull ghcr.io/obermarclp/the-nom-database/backend:v1.0.0

# Latest stable
docker pull ghcr.io/obermarclp/the-nom-database/backend:latest

# Development
docker pull ghcr.io/obermarclp/the-nom-database/backend:develop

# Specific commit
docker pull ghcr.io/obermarclp/the-nom-database/backend:main-abc123
```

## Setting Up Branch Protection

### On GitHub

1. Go to **Settings** → **Branches**
2. Add rule for `main`:
   - Require pull request reviews (1)
   - Require status checks: `test-backend`, `test-frontend`, `security-scan`
   - Require branches to be up to date
   - No force pushes
   - No deletions

3. Add rule for `develop`:
   - Require pull request reviews (1)
   - Require status checks: `test-backend`, `test-frontend`
   - Allow force pushes (maintainers only)

## Best Practices

### Do's ✅

- Keep feature branches small and focused
- Write descriptive commit messages
- Update documentation with code changes
- Write tests for new features
- Rebase feature branches before merging
- Delete branches after merging
- Tag releases consistently

### Don'ts ❌

- Never commit directly to main or develop
- Don't force push to shared branches
- Don't merge without PR review
- Don't commit secrets or credentials
- Don't skip CI checks
- Don't create giant PRs (split them up)
- Don't forget to update CHANGELOG

## Troubleshooting

### Accidentally Committed to Main

```bash
# Reset main to previous commit
git reset --hard HEAD~1

# Push force (if you have permission)
git push origin main --force

# Or create a revert commit
git revert HEAD
git push origin main
```

### Merge Conflicts

```bash
# Update your branch
git checkout feature/my-feature
git fetch origin
git merge origin/develop

# Resolve conflicts in files
# Then commit
git add .
git commit -m "chore: resolve merge conflicts"
```

### Forgot to Branch from Develop

```bash
# Stash your changes
git stash

# Switch to develop and create branch
git checkout develop
git pull origin develop
git checkout -b feature/my-feature

# Apply stashed changes
git stash pop
```

## Resources

- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

## Questions?

If you have questions about the workflow, please:
- Check this documentation first
- Ask in GitHub Discussions
- Contact the maintainers
