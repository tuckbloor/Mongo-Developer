# Mongo Developer V1.1 Query Workbench

V1.1 starts the move from a simple MongoDB browser toward a full developer database IDE.

## What changed

The old 62px query strip has been replaced with a large editor/results workspace.

- Find JSON/EJSON mode
- Aggregation Pipeline mode
- SQL convenience mode
- inline validation
- JSON formatting
- Ctrl+Enter execution
- query history
- draggable query/results split
- Expand/Restore for either pane
- Table and JSON result views
- clear zero-result states
- result copy
- selectable page size
- query timing
- `$lookup` starter
- SQL JOIN starter

## SQL mode

SQL mode is intentionally a convenience layer, not a second database engine. Supported V1.1 forms are translated to native MongoDB aggregation pipelines and executed by MongoDB.

Supported:

- `SELECT *`
- named fields and `AS` aliases
- `FROM collection`
- `FROM database.collection`
- base collection aliases
- `WHERE` conditions joined by `AND`
- `=`, `!=`, `>`, `>=`, `<`, `<=`
- `ORDER BY`
- `LIMIT`
- `JOIN`, `INNER JOIN`, `LEFT JOIN`

Example:

```sql
SELECT d.name, t.name AS team
FROM developers d
LEFT JOIN teams t ON d.team_id = t._id
WHERE d.active = true
ORDER BY d.name ASC
LIMIT 50;
```

The join becomes MongoDB `$lookup` followed by `$unwind`.

## Next competitive milestones

V1.2+ should focus on features that materially improve daily developer work rather than feature-count padding:

1. Explain-plan visualisation and index recommendations.
2. Query profiler and slow-query history.
3. Stage-by-stage aggregation builder with live intermediate output.
4. Schema distribution analysis, validation-rule authoring and schema drift comparison.
5. IntelliSense/autocomplete for fields, operators and collection names.
6. Ollama-powered natural-language query generation and query explanation, always previewed before execution.
7. Saved workspaces, query tabs, pinned queries and exportable query notebooks.
8. Data diff/sync between collections or environments with a read-only preview.
9. Transaction/session console for supported MongoDB deployments.
10. Native LexiBase adapter so the same GUI manages MongoDB and LexiBase without pretending LexiBase is MongoDB.

The quality bar is that each feature must be safer and easier to understand than the equivalent shell workflow, with destructive actions clearly separated and read-only profiles enforced by the Go backend.
