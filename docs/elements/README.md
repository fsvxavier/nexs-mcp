# NEXS MCP Elements Documentation

Complete documentation for all 6 element types in the NEXS MCP system.

## Element Types

### [Persona](PERSONA.md) 🎭
Defines AI agent personality, expertise, and communication style.

**Key Capabilities:**
- Behavioral traits with intensity levels
- Expertise areas with skill levels  
- Response style customization
- Privacy controls (public/private/shared)
- Hot-swappable

**Use Cases:** AI assistants, roleplay, specialized experts, customer service

---

### [Skill](SKILL.md) ⚡
Procedural capabilities triggered by conditions and executed as steps.

**Key Capabilities:**
- Trigger-based activation (keyword, pattern, context)
- Step-by-step procedures
- Tool integration
- Composable with dependencies

**Use Cases:** Code review, data pipelines, automation workflows, task execution

---

### [Template](TEMPLATE.md) 📝
Variable substitution and multi-format content generation.

**Key Capabilities:**
- {{variable}} syntax
- Multiple formats (Markdown, YAML, JSON, text)
- Required/optional variables
- Default values

**Use Cases:** Email templates, API responses, reports, documentation

---

### [Agent](AGENT.md) 🤖
Goal-oriented workflow execution with decision-making.

**Key Capabilities:**
- Multi-step action orchestration
- Decision trees
- Error recovery and fallback
- Context accumulation

**Use Cases:** Customer support, data analysis, research, automated workflows

---

### [Memory](MEMORY.md) 🧠
Persistent context storage with deduplication.

**Key Capabilities:**
- Text-based YAML storage
- Date-based organization
- SHA-256 content hashing
- Search indexing
- Custom metadata

**Use Cases:** Meeting notes, decisions, knowledge base, context preservation

---

### [Ensemble](ENSEMBLE.md) 🎼
Multi-agent orchestration with parallel execution.

**Key Capabilities:**
- Agent coordination
- Sequential/parallel/hybrid execution
- Result aggregation
- Fallback chains
- Shared context

**Use Cases:** Code review teams, research teams, multi-perspective analysis

---

## Quick Reference

| Element | Primary Use | Complexity | Dependencies |
|---------|-------------|------------|--------------|
| Persona | AI personality | Low | None |
| Skill | Task execution | Medium | Tools, other Skills |
| Template | Content generation | Low | None |
| Agent | Workflow orchestration | High | Skills, Tools |
| Memory | Context storage | Low | None |
| Ensemble | Multi-agent coordination | High | Agents |

## Element Relationships

```
Ensemble
  └─> Agent
       ├─> Skill
       │    └─> Template
       └─> Persona

Memory (standalone, referenced by any element)
```

## Common Patterns

### 1. Persona + Skill
A persona uses specific skills for task execution:
```
Technical Expert Persona → Code Review Skill → Template (report)
```

### 2. Agent + Multiple Skills
An agent orchestrates multiple skills:
```
Research Agent → [Search Skill, Analysis Skill, Summary Skill]
```

### 3. Ensemble Workflow
Multiple specialized agents:
```
Code Review Ensemble → [Security Agent, Performance Agent, Style Agent]
```

## Getting Started

1. **Create a Persona** for your use case
2. **Add Skills** for specific tasks
3. **Use Templates** for structured outputs
4. **Build Agents** for complex workflows
5. **Create Memories** to preserve context
6. **Orchestrate with Ensembles** for multi-agent tasks

## See Also

- [MCP Tools Reference](../TOOLS_SPEC.md)
- [Architecture](../plano/ARCHITECTURE.md)
- [Examples](../../examples/)
