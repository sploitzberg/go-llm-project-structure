# Mosaic Workflow Guide

This document explains how Mosaic is used as an external workflow tool for AI agents working on this project.

## Overview

Mosaic (HexxlaDB) is an external long-term memory and context retrieval system used by AI agents during development. It is NOT part of the project architecture - it's a separate tool used in the workflow.

Mosaic provides:

- Persistent storage of conversation turns
- Semantic search capabilities
- Structured querying with tags and filters
- Lattice-based context expansion
- Retention policy enforcement

## Relationship to Project

**Important:** Mosaic is NOT integrated into the project architecture. It is an external tool used by AI agents to:

- Store conversation history for context retrieval
- Maintain knowledge across sessions
- Enforce systematic patterns in agent behavior
- Provide persistent memory for the development workflow

The project itself (MCP Ratchet) follows strict Hexagonal Architecture independently. Mosaic operates outside this architecture as a workflow aid.

## Retention Policy

### Why save_all_turns?

The workflow uses `save_all_turns` capture mode to ensure:

- Complete audit trail of all AI interactions
- Ability to reconstruct conversation context
- Persistent knowledge across development sessions
- Systematic compliance with workflow requirements

### Enforcement Mechanism

The MCP server enforces the retention policy by:

- Returning errors when `put_cell` kinds conflict with `capture_mode`
- Rejecting operations that don't follow the required pattern
- Ensuring systematic compliance across all sessions

## Agent Workflow

When an AI agent works on this project:

1. **Initialization**
   - Load Mosaic configuration from `config/config.yaml`
   - Verify retention policy settings
   - Check database health with `mosaic_hexxla_health()`

2. **Per-Turn Processing**
   - Receive user message
   - Save to Mosaic: `put_cell(kind=user_message, ...)`
   - Retrieve relevant context using intelligent read patterns
   - Generate response
   - Save to Mosaic: `put_cell(kind=assistant_response, ...)`
   - Display response to user

3. **Context Retrieval**
   - Use `search_embedding` for semantic discovery
   - Use `query_cells` for structured queries
   - Use `load_context_pack` for lattice expansion
   - Apply appropriate tags and filters

4. **Knowledge Persistence**
   - Before writing, always search existing content
   - Check tag vocabulary with `list_tags` and `tag_counts`
   - Decide: reuse, supersede, or create new
   - Use atomic, specific tags

## MCP Ratchet Integration

### Ratchet Enforcement

MCP Ratchet enforces correct operation ordering:

- Requires `list_tags` before `put_cell` to prevent tag proliferation
- Ensures preparation steps are completed before writes
- Maintains data quality through systematic patterns

### Ratchet Tokens

The ratchet system uses one-time tokens:

- Token issued after completing required prerequisite
- Token must be provided to subsequent operation
- Token expires after use, requiring new cycle

This ensures agents follow the correct workflow every time.

## Configuration

### Mosaic Config Location

- **Config file**: `config/config.yaml`
- **Policy source**: Loaded at MCP server startup
- **Key settings**: capture_mode, enforcement, retention notes

### Environment Variables

Mosaic behavior can be controlled via environment variables (if configured):

- `MOSAIC_DB_PATH` — Database file location
- `MOSAIC_EMBED_MODEL` — Embedding model for semantic search
- `MOSAIC_POLICY_FILE` — Retention policy configuration

## Best Practices

### For Agents

- Always follow the retention policy systematically
- Use intelligent read/write patterns
- Prefer existing tags over creating new ones
- Load context from search hits, not arbitrary coordinates
- Verify compliance periodically

### For Developers

- Keep retention policy enforcement enabled
- Monitor database health and statistics
- Review tag counts for vocabulary hygiene
- Use verification tools to ensure compliance
- Document any policy changes in `docs/mosaic_retention_compliance.md`

### For Maintainers

- Regularly check `mosaic_hexxla_health()` output
- Monitor disk usage and database size
- Review retention policy settings periodically
- Update documentation when patterns change

## Troubleshooting

### Common Issues

**Issue**: put_cell fails with enforcement error

- **Cause**: Kind doesn't match capture_mode
- **Solution**: Ensure kind is `user_message` or `assistant_response` for `save_all_turns`

**Issue**: Can't find relevant content

- **Cause**: Using wrong search tool or no content exists
- **Solution**: Try `search_embedding` for semantic discovery, or content may not exist yet

**Issue**: Tags are inconsistent

- **Cause**: Not checking existing vocabulary before creating tags
- **Solution**: Always call `list_tags` and `tag_counts` before creating new tags

**Issue**: Context is missing neighboring information

- **Cause**: Not using `load_context_pack` with seeds
- **Solution**: After search, use `load_context_pack` with coordinates from hits

## References

- `docs/mosaic/AGENT_QUICK_REFERENCE.md` — Quick reference for agents
- `docs/mosaic/COMPLIANCE_CHECKLIST.md` — Compliance checklist
- `docs/mosaic_retention_compliance.md` — Retention policy documentation
- `docs/mcp-ratchet/OVERVIEW.md` — MCP Ratchet overview
- `.windsurf/rules/mosaic_mcp.md` — MCP agent workflow
