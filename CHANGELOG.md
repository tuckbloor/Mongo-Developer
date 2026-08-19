# V1.5.0

- Added Explain Query / Index Advice for Find JSON, Aggregation, SQL and Eloquent.
- Added Saved Queries with namespace/mode persistence and local deletion.
- Expanded query history to 50 entries with database, collection and duration metadata.
- Added Diagnostics with Copy Diagnostics and password-free connection details.
- Added connection health/latency/MongoDB-version indicator and Reconnect.
- Added unsaved query/document protection before connection/collection changes.
- Improved query errors with mode/namespace context, Copy Error and JSON line/column hints.
- Added Ctrl+L, Ctrl+Shift+F, Ctrl+S and broader Escape handling while retaining Ctrl+Enter/Ctrl+K.
- Hardened read-only mode against aggregation `$out` / `$merge` write stages in both frontend and backend.
- Retains V1.4 Backup & Restore and all Laravel/Eloquent/query-mode fixes.

# V1.4.0

- Added top-level Backup & Restore manager.
- Backup all user databases or one selected database.
- Timestamped ZIPs are retained in Mongo Developer's local backup store and a Save As flow offers a second downloaded/saved copy.
- Preserves BSON documents, collection options, views and index definitions.
- Restore replaces only databases contained in the archive and leaves unrelated databases untouched.
- Added restore from an external ZIP.
- Added stored backup listing, Download Copy, Restore and Delete ZIP actions.
- Backup creation is atomic via `.partial` files.
- Protected MongoDB system databases (`admin`, `config`, `local`) are excluded and cannot be restored.
- Restore is blocked on read-only profiles and requires destructive confirmation.
- Retains the V1.3.2 query-mode Run-state fix.

# V1.3.1

- Removed the SQL JOIN Example toolbar button and its example-population handler.
- SQL JOIN parsing/execution support remains available for manually written queries.
- Corrected Wails Windows version metadata and ensured `build/windows/info.json` is UTF-8 without BOM.

# Mongo Developer V1.2.2

- Compact five-line SQL editor by default.
- SQL editor auto-grows and shrinks with edited query line count, with a safe height cap.
- Old V1.2 editor-height preference no longer forces the query pane large.
- Keeps V1.2.1 simple SQL editor and JSON save validation.

# Mongo Developer V1.2.1

- Restored clean auto-loaded SQL text editor on collection click.
- Removed visible section-by-section SQL editor.
- Added live document JSON validation and save blocking for invalid JSON.
- Retained V1.2.0 table/document editing features.

# Mongo Developer V1.2.0

- Collection clicks now auto-generate and execute `SELECT * FROM <collection> LIMIT 50`.
- Added section-by-section SQL editing for SELECT, FROM, JOIN, WHERE, ORDER BY and LIMIT.
- Added explicit query dirty state: results stay unchanged until **Run Changes**.
- Added Raw SQL / Section Editor switching.
- Added table column resizing, visibility controls, pinned `_id`, and sticky row actions.
- Added inline editing for simple scalar cells with Save/Cancel controls.
- Added a right-side Fields/JSON document drawer for full-document editing and nested data.
- Added per-row View JSON, Edit, Duplicate and Delete actions.
- Editing is automatically disabled for transformed results without `_id` and for aggregation/read-only results.
- Added safe Refresh behavior that re-runs the last executed SQL while preserving unexecuted edits.
- Added structured SQL parser/builder tests and SQL-to-Mongo integration smoke checks.
- Retains V1.1.2 `dummy_test_data` one-click install/reset practice database.

# Mongo Developer V1.1.2

- Added built-in resettable `dummy_test_data` practice database.
- First writable local connection installs the practice database automatically when missing.
- Added one-click Install / Reset control in the Databases pane.
- Reset is hard-scoped to `dummy_test_data` only.
- Added 17 related collections, ~350 documents, nested fields, arrays, dates, nulls and useful indexes.
- Remote connections are never modified automatically; read-only profiles cannot install/reset practice data.

# Changelog

## 1.1.1 - 2026-08-19

- Fixed Windows release-gate failure where Vitest discovered `src/sql.node.test.mjs` even though it was a standalone Node `node:test` smoke test.
- Renamed the standalone smoke test to `src/sql.node.smoke.mjs`, outside Vitest's `*.test.*` discovery pattern.
- Kept the real SQL translator Vitest coverage in `src/sql.test.js` unchanged.
- Updated the `test:sql` npm script and installer version defaults to 1.1.1.

## 1.1.0 - 2026-08-19

- Replaced the cramped single-line query strip with the Query Workbench.
- Added Find JSON, Aggregation and SQL query modes.
- Added SQL-to-Mongo aggregation translation for SELECT, AND-based WHERE comparisons, ORDER BY, LIMIT, JOIN and LEFT JOIN.
- Added built-in `$lookup` and SQL JOIN examples.
- Added inline query validation and JSON formatting.
- Added Ctrl+Enter query execution.
- Added draggable query/results splitter with persisted editor height.
- Added Expand/Restore controls for both the query and result areas.
- Added unified Table/JSON results, removing the confusing empty standalone JSON view.
- Added query history (25 most recent distinct queries).
- Added result copy action, clearer zero-result messaging and selectable result page sizes.
- Updated default local URI to include the trailing slash (`mongodb://localhost:27017/`).
- Added standalone Node tests for the SQL translator as well as Vitest coverage.

## 1.0.0 - 2026-08-18

- Saved MongoDB connections with Windows DPAPI password protection.
- Local MongoDB and Atlas-compatible URI support.
- Database and collection explorer.
- Create/drop database and collection operations.
- Document browser with EJSON-aware filters, sorting and pagination.
- Insert, replace and delete document operations.
- JSON view and observed-schema view.
- Collection statistics.
- Index listing and index creation.
- Aggregation pipeline execution.
- JSON import/export.
- Per-connection read-only safety mode.
- Laravel project inspection and Eloquent MongoDB model generation.
- Windows Wails/NSIS installer build pipeline.
- GitHub CI and tagged-release workflow.
