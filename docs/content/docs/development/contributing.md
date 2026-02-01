---
title: "Contributing"
description: "How to contribute to The Nom Database"
weight: 410
icon: "code"
toc: true
---

# Contributing to The Nom Database

Thank you for your interest in contributing to The Nom Database! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our code of conduct:

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on what is best for the community
- Show empathy towards other community members

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates.

**Good bug reports include:**

- **Clear title** describing the problem
- **Steps to reproduce** the behavior
- **Expected behavior** vs actual behavior
- **Screenshots** if applicable
- **Environment details** (OS, Docker version, etc.)
- **Logs** or error messages

**Bug Report Template:**

```markdown
**Description:**
A clear description of the bug.

**Steps to Reproduce:**
1. Go to '...'
2. Click on '...'
3. See error

**Expected Behavior:**
What should happen.

**Actual Behavior:**
What actually happens.

**Environment:**
- OS: [e.g., macOS 13.0, Ubuntu 22.04]
- Docker version: [e.g., 24.0.0]
- Browser: [e.g., Chrome 120]

**Logs:**
```
Paste relevant logs here
```

**Screenshots:**
If applicable, add screenshots.
```

### Suggesting Features

Feature suggestions are welcome! Please:

- **Check existing issues** to avoid duplicates
- **Describe the problem** your feature would solve
- **Explain your proposed solution**
- **Consider alternatives** you've thought about

**Feature Request Template:**

```markdown
**Problem:**
Describe the problem or limitation.

**Proposed Solution:**
How would you solve this?

**Alternatives:**
What other solutions have you considered?

**Additional Context:**
Any other relevant information.
```

### Contributing Code

#### First Time Contributors

Look for issues labeled `good first issue` or `help wanted`.

#### Development Setup

1. **Fork the repository**
   ```bash
   # Click "Fork" on GitHub
   ```

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/the-nom-database.git
   cd the-nom-database
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/the-nom-database.git
   ```

4. **Create development environment**
   ```bash
   # Copy environment file
   cp .env.example .env

   # Edit configuration
   nano .env

   # Start services
   docker compose up -d
   ```

5. **Verify setup**
   ```bash
   # Check services
   docker compose ps

   # Open application
   open http://localhost:3000
   ```

#### Making Changes

1. **Create a branch**
   ```bash
   # Update develop branch
   git checkout develop
   git pull upstream develop

   # Create feature branch
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow code style guidelines
   - Add tests for new features
   - Update documentation
   - Keep commits focused and atomic

3. **Test your changes**

   **Backend tests:**
   ```bash
   cd backend
   go test ./...
   go test -race ./...
   go vet ./...
   ```

   **Frontend tests:**
   ```bash
   cd frontend
   npm test
   npm run lint
   npm run build
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` New feature
   - `fix:` Bug fix
   - `docs:` Documentation
   - `test:` Tests
   - `refactor:` Code refactoring
   - `style:` Formatting
   - `chore:` Maintenance

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create Pull Request**
   - Go to GitHub and create a PR
   - Fill out the PR template
   - Link related issues
   - Request reviews

#### Pull Request Guidelines

**Before submitting:**

- [ ] Code follows project style
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts
- [ ] Self-reviewed the code

**PR Template:**

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Related Issues
Fixes #123
Related to #456

## How Has This Been Tested?
Describe the tests you ran.

## Screenshots (if applicable)
Add screenshots for UI changes.

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review
- [ ] I have commented my code where necessary
- [ ] I have updated the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix/feature works
- [ ] New and existing tests pass locally
```

## Code Style Guidelines

### Go (Backend)

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused
- Add comments for exported functions
- Handle errors properly

**Example:**

```go
// GetRestaurant retrieves a restaurant by ID
func GetRestaurant(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
        return
    }

    // ... implementation
}
```

### TypeScript/React (Frontend)

- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Keep components small and reusable
- Use Tailwind CSS for styling
- Add prop types/interfaces

**Example:**

```typescript
interface RestaurantCardProps {
  restaurant: Restaurant;
  onEdit?: () => void;
  onDelete?: () => void;
}

export const RestaurantCard: React.FC<RestaurantCardProps> = ({
  restaurant,
  onEdit,
  onDelete,
}) => {
  // ... implementation
};
```

### General Guidelines

- **DRY**: Don't Repeat Yourself
- **KISS**: Keep It Simple, Stupid
- **YAGNI**: You Aren't Gonna Need It
- Write self-documenting code
- Add comments for complex logic
- Use meaningful variable names
- Keep lines under 100 characters
- Add blank lines for readability

## Testing Guidelines

### Backend Tests

```go
func TestGetRestaurant(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()

    // Test case
    restaurant := createTestRestaurant(t, db)

    // Execute
    result, err := GetRestaurantByID(db, restaurant.ID)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, restaurant.Name, result.Name)
}
```

### Frontend Tests

```typescript
describe('RestaurantCard', () => {
  it('renders restaurant information', () => {
    const restaurant = createMockRestaurant();
    render(<RestaurantCard restaurant={restaurant} />);

    expect(screen.getByText(restaurant.name)).toBeInTheDocument();
    expect(screen.getByText(restaurant.address)).toBeInTheDocument();
  });
});
```

### Test Requirements

- **Coverage**: Aim for >80% code coverage
- **Unit tests**: Test individual functions
- **Integration tests**: Test component interactions
- **E2E tests**: Test critical user flows (optional)

## Documentation

### Code Documentation

- Add comments for exported functions
- Document complex algorithms
- Explain "why" not just "what"
- Keep documentation up to date

### User Documentation

- Update README.md if needed
- Add to docs/ folder for major features
- Include examples and screenshots
- Keep language clear and simple

## Review Process

### What Reviewers Look For

1. **Correctness**: Does it work as intended?
2. **Tests**: Are there adequate tests?
3. **Performance**: Any performance concerns?
4. **Security**: Any security issues?
5. **Style**: Follows code style guidelines?
6. **Documentation**: Is it well documented?

### Responding to Reviews

- Be open to feedback
- Ask questions if unclear
- Make requested changes
- Mark conversations as resolved
- Thank reviewers for their time

### Merging

Once approved and CI passes:
1. **Maintainer** will merge the PR
2. **Feature branch** will be deleted
3. **Docker images** will be rebuilt (if on main/develop)

## Release Process

See [Git Workflow Guide](/docs/development/git-workflow/) for detailed release process.

**Quick overview:**
1. Features merged to `develop`
2. Release branch created from `develop`
3. Testing and bug fixes
4. PR to `main` branch
5. Tag release (triggers automation)
6. Docker images published automatically

## Getting Help

### Resources

- **Documentation**: [https://obermarclp.github.io/the-nom-database](https://obermarclp.github.io/the-nom-database)
- **Git Workflow**: [Git Workflow](/docs/development/git-workflow/)
- **API Docs**: [API Reference](/docs/configuration/api/)

### Ask Questions

- **GitHub Discussions**: For general questions
- **GitHub Issues**: For bugs and features
- **Pull Request Comments**: For code-specific questions

### Community

- Be patient and respectful
- Help others when you can
- Share your knowledge
- Welcome newcomers

## Recognition

Contributors are recognized in:
- GitHub contributors page
- Release notes
- CHANGELOG.md

## License

By contributing, you agree that your contributions will be licensed under the BSD 3-Clause License.

## Thank You!

Your contributions make this project better for everyone. Thank you for taking the time to contribute!

---

**Questions?** Open a [Discussion](https://github.com/obermarclp/the-nom-database/discussions) or reach out to maintainers.
