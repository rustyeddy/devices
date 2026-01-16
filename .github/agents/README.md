# Custom Agents Directory

This directory is reserved for custom GitHub Copilot agent definitions.

## Purpose
Custom agents allow you to create specialized AI assistants with focused expertise for specific tasks or domains within this repository.

## How to Add Custom Agents
Each custom agent should be defined in its own file with the `.agent.md` extension.

### Example Structure
```markdown
---
name: test_agent
description: Specialized agent for writing and maintaining tests
applies_to: "**/*_test.go"
---

## Responsibilities
- Write comprehensive tests following testify patterns
- Ensure hermetic test design
- Focus on edge cases and error paths

## Boundaries
- Only modify test files
- Never change production code
```

## Current Agents
Currently, no custom agents are defined. The repository uses:
- **Repository-wide instructions**: `.github/copilot-instructions.md`
- **Quick reference**: `AGENTS.md` at repository root

## When to Add Custom Agents
Consider adding custom agents for:
- Specialized testing domains (e.g., hardware mocks, sensor testing)
- Documentation maintenance
- Specific driver families that need domain expertise
- Security-focused code reviews

## Documentation
- [GitHub Copilot Custom Agents](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/create-custom-agents)
- [Best Practices for AGENTS.md](https://github.blog/ai-and-ml/github-copilot/how-to-write-a-great-agents-md-lessons-from-over-2500-repositories/)
