# Pull Request

## Description
<!-- Provide a clear and concise description of your changes -->

## Type of Change
<!-- Mark the relevant option with an 'x' -->

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring
- [ ] Test improvements
- [ ] CI/CD improvements
- [ ] Element submission (Persona, Skill, Template, Agent, Memory, Ensemble)

## Related Issues
<!-- Link to related issues using #issue_number -->

Closes #
Relates to #

## Changes Made
<!-- List the specific changes you made -->

- 
- 
- 

## Element Submission (if applicable)
<!-- Complete this section only if submitting a new element -->

### Element Details
- **Element Type:** (Persona/Skill/Template/Agent/Memory/Ensemble)
- **Element Name:** 
- **Element ID:** 
- **Category:** 
- **Tags:** 

### Element Validation
```shell
# Paste the output of: nexs-mcp validate-element <your-element>.yaml
```

## Testing
<!-- Describe the tests you ran and their results -->

### Test Environment
- **OS:** (macOS/Linux/Windows)
- **Go Version:** 
- **NEXS-MCP Version:** 

### Tests Performed
- [ ] All existing tests pass (`make test`)
- [ ] Added new tests for new functionality
- [ ] Tested manually with Claude Desktop
- [ ] Tested edge cases
- [ ] Tested error handling
- [ ] Performance tested (if applicable)

### Test Results
```shell
# Paste relevant test output
```

## Documentation
<!-- Have you updated the documentation? -->

- [ ] Updated README.md (if needed)
- [ ] Updated relevant documentation in `docs/`
- [ ] Added/updated code comments
- [ ] Updated CHANGELOG.md
- [ ] Added usage examples

## Code Quality
<!-- Ensure your code meets quality standards -->

- [ ] Code follows Go best practices
- [ ] Code follows project coding standards (see CONTRIBUTING.md)
- [ ] Ran `go fmt` and `go vet`
- [ ] No new linter warnings (`make lint`)
- [ ] Added appropriate error handling
- [ ] Added logging where appropriate
- [ ] No hardcoded credentials or sensitive data

## Performance Impact
<!-- Describe any performance implications -->

- [ ] No significant performance impact
- [ ] Performance improved
- [ ] Performance degraded (explain why this is acceptable)

### Performance Details
<!-- If performance changed, provide details -->

## Breaking Changes
<!-- If this is a breaking change, describe the impact and migration path -->

### Impact

### Migration Path

## Checklist
<!-- Ensure all items are complete before submitting -->

- [ ] My code follows the project's coding standards
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published
- [ ] I have checked my code and corrected any misspellings

## Screenshots (if applicable)
<!-- Add screenshots to help explain your changes -->

## Additional Context
<!-- Add any other context about the PR here -->

## Reviewer Notes
<!-- Specific areas you'd like reviewers to focus on -->

---

## For Maintainers

### Review Checklist
- [ ] Code quality is acceptable
- [ ] Tests are comprehensive
- [ ] Documentation is complete
- [ ] No security issues
- [ ] No performance regressions
- [ ] Breaking changes are documented
- [ ] CHANGELOG.md is updated

### Merge Strategy
- [ ] Squash and merge
- [ ] Merge commit
- [ ] Rebase and merge
