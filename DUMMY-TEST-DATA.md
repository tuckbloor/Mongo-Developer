# Mongo Developer — Built-in Dummy Test Data

Mongo Developer V1.2.1 includes the resettable practice database introduced in V1.1.2:

```text
dummy_test_data
```

## Behaviour

- On the first successful **local writable** connection (`localhost`, `127.0.0.1`, or `::1`), Mongo Developer checks for `dummy_test_data` and creates it if it is missing.
- Remote connections are never modified automatically.
- Read-only profiles never install or reset practice data.
- The Databases pane always shows **Install** when the practice database is absent and **Reset** when it exists.
- Reset drops and recreates **only** `dummy_test_data`. No other database is touched.

## Included collections

| Collection | Purpose |
|---|---|
| departments | Parent departments for team joins |
| teams | Teams linked to departments |
| developers | People, roles, skills, nested profile/address data |
| customers | Companies with nested contacts and addresses |
| suppliers | Product suppliers |
| categories | Product categories |
| products | Products with prices, tags and nested specs |
| warehouses | Warehouse reference data |
| inventory | Product and warehouse stock joins |
| projects | Projects linked to customers and developers |
| project_members | Many-to-many project/developer practice |
| orders | Orders with nested shipping and status history |
| order_items | Order/product line items |
| support_tickets | Tickets linked to customers and developers |
| events | Time-based event data |
| audit_logs | Audit/event style records |
| settings | Mixed BSON types including null, arrays and nested objects |

The seed contains approximately 350 documents plus useful unique, foreign-key-style and compound indexes.

## Practice queries

### Find every developer

```json
{}
```

### Find active developers

```json
{ "active": true }
```

### Find nested profile values

```json
{ "profile.theme": "dark" }
```

### Aggregation join — developers to teams

```json
[
  {
    "$lookup": {
      "from": "teams",
      "localField": "team_id",
      "foreignField": "_id",
      "as": "team"
    }
  },
  { "$unwind": "$team" }
]
```

### SQL mode join

```sql
SELECT d.name, d.role, t.name AS team_name
FROM developers d
LEFT JOIN teams t ON d.team_id = t._id
WHERE d.active = true
ORDER BY d.name ASC
LIMIT 50;
```

### SQL mode — orders and customers

```sql
SELECT o.order_no, o.status, c.company
FROM orders o
LEFT JOIN customers c ON o.customer_id = c._id
ORDER BY o.created_at DESC
LIMIT 50;
```

## Safety

The reset path uses the constant database name `dummy_test_data`. The UI confirmation explicitly states that only this database is dropped and rebuilt. The general MongoDB database-drop controls remain separate.
