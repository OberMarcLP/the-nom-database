+++
title = "Testing"
description = "Backend and frontend testing guide"
weight = 440
icon = "science"
toc = true
+++

# Testing Guide

This project includes comprehensive automated tests for both the backend (Go) and frontend (React/TypeScript).

## Running Tests

### Run All Tests
```bash
make test
```

### Backend Tests Only
```bash
make test-backend
# Or manually:
cd backend && go test -v ./...
```

### Frontend Tests Only
```bash
make test-frontend
# Or manually:
cd frontend && npm test
```

### Test Coverage
```bash
make test-coverage
```

This generates coverage reports:
- Backend: `backend/coverage.html`
- Frontend: `frontend/coverage/index.html`

## Backend Tests

### Test Structure
- **Location**: `backend/internal/*/`
- **Framework**: Go's built-in testing package
- **Naming**: `*_test.go`

### Test Coverage

#### Image Processing (`internal/services/imageprocessor_test.go`)
- Image resizing with aspect ratio preservation
- Thumbnail generation
- JPEG compression
- Handling different image sizes (small, large, square)

#### Restaurant Handlers (`internal/handlers/restaurants_test.go`)
- Input validation
- Filter query building
- Restaurant data validation

### Writing Backend Tests

```go
func TestYourFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"test case 1", "input1", "expected1"},
        {"test case 2", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Frontend Tests

### Test Structure
- **Location**: `frontend/src/components/` (next to components)
- **Framework**: Vitest + React Testing Library
- **Naming**: `*.test.tsx`
- **Setup**: `frontend/src/test/setup.ts`

### Test Coverage

#### StarRating Component
- Rendering correct number of stars
- Display rating values
- Interactive mode (onChange callbacks)
- Readonly mode
- Half ratings support

#### ThemeToggle Component
- Toggle button rendering
- Icon switching (sun/moon)
- Click handlers
- Accessibility (aria-labels)

#### LazyImage Component
- Intersection Observer integration
- Loading skeleton display
- Image lazy loading
- Error handling

### Writing Frontend Tests

```typescript
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { YourComponent } from './YourComponent';

describe('YourComponent', () => {
  it('renders correctly', () => {
    render(<YourComponent />);
    expect(screen.getByText('Expected Text')).toBeInTheDocument();
  });

  it('handles user interaction', () => {
    const handleClick = vi.fn();
    render(<YourComponent onClick={handleClick} />);

    fireEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalled();
  });
});
```

## Configuration

### Backend
No additional configuration needed - uses Go's built-in testing framework.

### Frontend (`frontend/vitest.config.ts`)
- **Environment**: happy-dom (lightweight DOM implementation)
- **Globals**: Enabled for describe/it/expect
- **Setup Files**: `src/test/setup.ts`
- **Coverage Provider**: v8

## Mocks

### Frontend Mocks (in `setup.ts`)
- `window.matchMedia` - For responsive design testing
- `IntersectionObserver` - For lazy loading components
- Google Maps API (if needed for map components)

## Best Practices

1. **Test Behavior, Not Implementation**
   - Focus on what the component/function does, not how
   - Test user interactions and visible outcomes

2. **Use Descriptive Test Names**
   - Good: `"should display error when file size exceeds limit"`
   - Bad: `"test validation"`

3. **Arrange-Act-Assert Pattern**
   ```typescript
   it('example test', () => {
     // Arrange: Set up test data
     const input = 'test';

     // Act: Execute the code being tested
     const result = myFunction(input);

     // Assert: Verify the outcome
     expect(result).toBe('expected');
   });
   ```

4. **Test Edge Cases**
   - Empty inputs
   - Maximum values
   - Null/undefined
   - Error conditions

5. **Keep Tests Independent**
   - Each test should run in isolation
   - No shared state between tests
   - Use beforeEach/afterEach for setup/cleanup

## Continuous Integration

Tests should be run:
- Before committing code
- In CI/CD pipeline
- Before merging pull requests

## Adding New Tests

When adding new features:
1. Write tests alongside the feature code
2. Aim for at least 70% code coverage
3. Test happy paths and error cases
4. Update this document if adding new test patterns

## Troubleshooting

### Frontend Tests Fail to Import Components
- Check that `vitest.config.ts` has the correct path alias
- Ensure `tsconfig.json` includes test files

### Backend Tests Don't Run
- Verify test files end with `_test.go`
- Check that you're in the correct directory
- Run `go mod tidy` to ensure dependencies are installed

### Coverage Reports Not Generated
- Ensure you have write permissions in the project directory
- Check that coverage tools are installed (v8 for frontend)
