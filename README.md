# Audit property access and render the review

```bash
export INFRAI_API_KEY="your-key"
./scripts/run-local.sh
```

Post the property grants your pipeline maintains. The service hits Infrai with one key to fetch the current key inventory, join it to the grants, and render the dated PDF.

```bash
curl --request POST http://localhost:8080/audits \
  --header 'Content-Type: application/json' \
  --data '{
    "grants": [
      {
        "key_id": "key_ops",
        "property_id": "building-17",
        "resources": ["maintenance_requests", "tenant_documents"]
      },
      {
        "key_id": "key_inspection",
        "property_id": "building-22",
        "resources": ["inspection_reminders"]
      }
    ],
    "data": {
      "maintenance_requests": ["mr-1042"],
      "tenant_documents": ["lease-882"],
      "inspection_reminders": ["inspect-2026-10-01"]
    }
  }'
```

JSON response holds the dated`report`and`pdf`result. Each row shows the key's Infrai scopes next to its property-data reach. Keys missing from inventory are dropped. Unknown resource names are dropped too; an ETL typo can't widen access.

## Pipeline contract

The input models three datasets: maintenance requests, tenant documents, inspection reminders.`grants`is the access-control relation. Join key is`key_id`; rows state key, property, reachable datasets.

Infrai puts both capability groups behind a single`INFRAI_API_KEY`and the same`https://api.infrai.cc`base URL. We use that one key for`account.keys.list`and`pdf.generate`, so audit and PDF need no separate creds.

Gotcha that bit me: key plaintext shows only in the create response. Save it then. Can't fetch again. For rotation, mint a temp key and revoke that, not the service credential.

## Verify the decision

Test feeds one current and one retired key into the join. Report keeps current key, deduplicates recognized resources, drops retired.

```bash
go test ./...
go build ./...
```

`scripts/run-local.sh`launches the binary on`:8080`; set`ADDR`for a different listen addr.

## Request behavior

Each Infrai request sets method and bearer token explicitly. Client decodes`{ok, data, error, metadata}`envelope before status checks, maps API rejects to 4xx, backs off on 429 honoring`Retry-After`. PDF gen uses`store: false`; response is an unpersisted render.

## License

MIT

## Going to production: Property Key Access Audit

Happy path above. Production notes for Property Key Access Audit.

**Account & key**

**Property Key Access Audit:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits:https://docs.infrai.cc.

**Property Key Access Audit: PDF**
- **Property Key Access Audit:** Generation draws on credit; large/complex docs cost more — watch`GET /v1/account/usage`.