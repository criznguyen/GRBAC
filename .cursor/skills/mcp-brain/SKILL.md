---
name: mcp-brain
description: Long-term memory for AI agents. Load when user asks about project knowledge, decisions, "what did we", "recall", "summarize", or when coordinating team work. Use memory recall before responding; ingest when new facts emerge.
---

# mcp-brain — Long-term Memory for AI Agents

Use this skill when the user needs persistent project memory: tech stack, decisions, conventions, handoffs, or team coordination.

## When to Load

- User asks: "what did we decide?", "recall", "summarize", project knowledge, tech stack
- Coordinating work: handoff, tasks, decisions, contracts
- Before implementing: recall architecture, API spec, DB schema

## 1. Recall Protocol

```
1. memory_get_default_namespace → namespace
2. memory_recall_compact(namespace, query=<user intent>, budget_tokens=300)
3. Use working_set_md as primary context — no @files if it answers
4. Same-chat follow-up: memory_recall_delta(last_pack_hash)
5. Detail needed: memory_expand(pointer_ids)
```

## 2. Ingest Protocol — MANDATORY

**Call at END of every turn** with substantive content. Do NOT skip.

- **Per turn:** `memory_chat_history_store` + `memory_ingest_turn_summary`
- **Session checkpoint (user says "compact"):** `memory_pre_compact(namespace, messages, summary)` — one call stores raw + summary
- **Structured**: `memory_ingest(namespace, items)` with DECISION/FACT/CONSTRAINT format

## 3. Extended Tools

| Tool | Use when |
|------|----------|
| memory_pre_compact | Session checkpoint before compact (user says "compact", context 50–70%) |
| memory_suggest_recall | Need related topics before responding |
| memory_analytics | Check recall hit rate, token usage |
| memory_resource_working_set | Get pack as resource |
| memory_recall_compact(..., as_of_ts) | Point-in-time: what we knew at T |
| memory_recall_new_since(since_ts) | Session start: what's new since last_session_end |
| memory_recall_headlines(query) | Headlines-only (~150 tokens); then expand IDs as needed |
| memory_recall_compact(..., entity_bundle=true) | Group facts/decisions by entity (PostgreSQL, Redis, etc.) |
| memory_recall_compact(..., timeline_collapse=true) | Timeline by date; timeline_by_date[].pointer_ids for expand |
| memory_recall_compact(..., granularity=macro) | Ultra-compact: macro (~50 tokens), medium, micro |
| memory_expand(..., granularity=medium\|full) | Override mode: 256 or full content |
| memory_ingest(..., items[].version) | Skill versioning for procedural |

## 4. Team Coordination

- `team_handoff_create` / `team_handoff_read` — handoff between roles
- `team_decision_propose` / `team_decision_accept` — decisions with evidence
- `team_contract_get` / `team_contract_set` — api-spec, db-schema
- `team_task_create` / `team_task_list` — tasks and blockers
- `team_digest` — standup/awareness since timestamp

## Token Rule

Recall replaces code reads. Do not add Cursor search or @files when recall has the answer.
