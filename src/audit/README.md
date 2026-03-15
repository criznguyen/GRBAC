# Audit Module

Writes check events and admin action events to Audit Store.

- Async write for check events (p99 < 50ms)
- Sync or async for admin events
- Append-only; no update/delete
