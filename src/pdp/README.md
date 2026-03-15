# Policy Decision Point (PDP)

Check Permission logic — resolves subject → roles → permissions, evaluates (resource, action), returns Allow/Deny.

- Integrates with Redis cache
- Async audit write (off critical path)
- REST: POST /check, POST /check/batch
