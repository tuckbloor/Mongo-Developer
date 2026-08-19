# Security Policy

## Credentials

Mongo Developer never intentionally writes MongoDB passwords into `connections.json`.
On Windows, stored passwords are encrypted using Windows DPAPI for the current user.
The URI retained in the connection profile has its password removed.

Read-only connection profiles block create, update, delete, drop, import, and index-creation operations in the application layer.
This is a safety feature, not a substitute for MongoDB server-side least-privilege users.

## Reporting a vulnerability

Please report security issues privately to the maintainers before opening a public issue.
Do not include real production credentials, connection strings, database dumps, or customer data in reports.
