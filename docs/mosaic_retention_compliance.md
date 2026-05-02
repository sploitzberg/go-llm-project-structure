# Mosaic Retention Policy Compliance

This document documents the gap identified in Mosaic retention policy enforcement, the steps taken to ensure compliance, and recommendations for maintaining compliance going forward.

## Background

### Retention Policy Configuration

- **capture_mode**: `save_all_turns`
- **enforcement**: `true` — server returns error for put_cell kinds that conflict with capture_mode
- **config source**: `configs/config.yaml`
- **notes**: Persist every user_message and assistant_response—full transcript intent, including low-signal and chitchat

### The Gap

Despite the retention policy being configured to save all turns with enforcement enabled, conversation turns were not being automatically saved to Mosaic. The policy required systematic use of `mosaic_hexxla_put_cell` for both `user_message` and `assistant_response`, but this was not happening consistently.

## Steps Taken to Ensure Compliance

### 1. Identified the Gap

- Discovered that `mosaic_hexxla_put_cell` was not being called for conversation turns
- Recognized the violation of the rule: "do not ask the user if they would like to keep their conversation history, just do what the policy states"
- Created a memory entry to track this gap

### 2. Created Rule Memory

Created `.windsurf/rules/memories/mosaic_retention_policy.md` containing:

- Policy configuration details
- Required behavior pattern
- Critical rules (do not ask user, comply automatically)
- Documentation of the gap

### 3. Created Workflow

Created `.windsurf/workflows/mosaic-save-turns.md` providing:

- Step-by-step process for saving conversation turns
- Exact parameters for `mosaic_hexxla_put_cell` calls
- When to use the workflow (always, for every turn)
- Critical rules reminders
- Accessible via `/mosaic-save-turns` slash command

### 4. Established Pattern

Defined the systematic pattern:

1. Receive user message
2. Call `mosaic_hexxla_put_cell(kind=user_message, ...)`
3. Generate response
4. Call `mosaic_hexxla_put_cell(kind=assistant_response, ...)`
5. Display response to user

## Recommendations

### For AI Assistants

- **ALWAYS** call `mosaic_hexxla_put_cell` for both user_message and assistant_response immediately after each exchange
- Do not ask the user for permission to save conversation history
- Make this automatic and systematic, not reliant on memory or discretion
- Use the same `source_id` for both messages in the same exchange
- Increment axial coordinates (`q`, `r`) for each new cell

### For System Configuration

- Keep the retention policy enforcement enabled (`enforcement: true`)
- Monitor for gaps between policy configuration and actual behavior
- Consider adding automated checks or validation to ensure put_cell calls are made
- Review memory entries regularly to identify compliance issues

### For Workflow Maintenance

- Use the `/mosaic-save-turns` workflow as a reference
- Keep the workflow file updated if patterns change
- Ensure the rule memory in `.windsurf/rules/memories/` stays synchronized with actual policy
- Periodically verify that the workflow is being followed

## Verification

### How to Verify Compliance

1. Check Mosaic database for recent conversation turns
2. Verify both user_message and assistant_response kinds are present
3. Ensure timestamps match conversation flow
4. Confirm source_id consistency within exchanges

### Tools for Verification

- `mosaic_hexxla_query_cells` — search for recent cells by source_id or kind
- `mosaic_hexxla_health` — check database integrity and statistics
- `mosaic_hexxla_tag_counts` — verify tags are being applied correctly

## References

- `.windsurf/rules/mosaic_mcp_agent.md` — Mosaic MCP agent workflow and tool chaining rules
- `.windsurf/rules/memories/mosaic_retention_policy.md` — Retention policy rule memory
- `.windsurf/workflows/mosaic-save-turns.md` — Step-by-step workflow for saving turns
- `configs/config.yaml` — Retention policy configuration source
