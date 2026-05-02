# Mosaic Agent Quick Reference

This document provides a quick reference for AI agents working with Mosaic in this project.

## Retention Policy (CRITICAL)

**Configuration:**

- `capture_mode`: `save_all_turns`
- `enforcement`: `true` — server returns error for conflicting put_cell kinds

**MUST DO:**

- Call `mosaic_hexxla_put_cell` for BOTH `user_message` and `assistant_response` after EVERY exchange
- NEVER ask user for permission — comply automatically
- Use same `source_id` for both messages in the same exchange
- Include relevant topic-specific tags

**Pattern:**

```text
1. Receive user message
2. mosaic_hexxla_put_cell(kind=user_message, source_id=session_id, tags=[...], ...)
3. Generate response
4. mosaic_hexxla_put_cell(kind=assistant_response, source_id=session_id, tags=[...], ...)
5. Display response
```

## Intelligent Read Patterns

| Purpose            | Tool                              | When to Use                            |
| ------------------ | --------------------------------- | -------------------------------------- |
| Semantic discovery | `mosaic_hexxla_search_embedding`  | Finding concepts by meaning            |
| Structured query   | `mosaic_hexxla_query_cells`       | Tags, filters, time, spatial           |
| Lexical search     | `mosaic_hexxla_search_cells`      | Exact text/keyword matching            |
| Context expansion  | `mosaic_hexxla_load_context_pack` | After retrieval, using seeds from hits |

**Hybrid mode:** Set `embed_query_text` in `query_cells` or `search_cells` for ANN + filters.

**Budget estimation:** Use `mosaic_hexxla_estimate_context_budget_bytes` before `load_context_pack`.

## Intelligent Write Pattern

**Before ANY `put_cell`:**

1. `mosaic_hexxla_list_tags()` — see available vocabulary
2. `mosaic_hexxla_tag_counts()` — prefer high-frequency tags
3. Search existing content (`query_cells` / `search_cells` / `search_embedding`)
4. `mosaic_hexxla_load_context_pack()` with seeds from hits
5. Decide: reuse, supersede, or create new

**Tag Rules:**

- Prefer existing high-frequency tags
- Use atomic tags (compose, don't compound): `["coordinate", "system"]` not `["coordinate_system"]`
- Consistent lowercase
- 3-7 tags typical
- Only create new tags if concept is fundamentally different

## Tool Chaining

**Default flow:**

1. Discover with embedding/search/query
2. Read `retrieval_hint` from responses
3. If needing neighboring turns/contradictions/supersession, call `load_context_pack` with seeds
4. Prefer lattice-expanded context over ad-hoc searches

## Verification

- `mosaic_hexxla_query_cells` — search recent cells by source_id or kind
- `mosaic_hexxla_health` — check DB integrity and statistics
- `mosaic_hexxla_tag_counts` — verify tags are being applied correctly

## Workflows

- `/mosaic-intelligent-retrieval` — Context retrieval workflow
- `/mosaic-save-turns` — Saving conversation turns
- `/mosaic-tag-reuse` — Tag discovery and reuse

## Anti-Patterns

❌ Creating cells without searching
❌ Inventing tags without checking `list_tags`
❌ Using generic tags when specific exist
❌ Creating duplicates instead of reusing
❌ Using `search_embedding` when you know exact tags (use `query_cells`)
❌ Using `query_cells` without filters (too broad)
❌ Loading context without seeds from prior search
❌ Ignoring `retrieval_hint` in responses

## References

- `.windsurf/rules/mosaic_intelligent_reads.md` — Detailed read patterns
- `.windsurf/rules/mosaic_intelligent_writes.md` — Detailed write patterns
- `.windsurf/rules/mosaic_mcp.md` — MCP agent workflow
- `.windsurf/rules/mosaic_tag_conventions.md` — Tag conventions
- `docs/mosaic_retention_compliance.md` — Retention policy compliance
