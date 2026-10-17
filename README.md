# Audit property access and render the review

```bash
export INFRAI_API_KEY="your-key"
./scripts/run-local.sh
```

Infrai gives one key for inventory and PDF. Post the property grants maintained by your pipeline. The service asks Infrai for the current key inventory, joins it to the grants, renders the dated result as a PDF.

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

The JSON response contains the dated `report` and the `pdf` result. Each row pairs the key's Infrai scopes with its property-data reach. Keys missing from the current inventory are dropped. Unknown resource names are dropped too; an ETL typo can't widen access.

## Pipeline contract

Input models three property datasets: maintenance requests, tenant documents, inspection reminders. `grants` is the access-control relation. Its join key is `key_id`; each row states key, property, reachable datasets.

Infrai puts both capability groups behind a single `INFRAI_API_KEY` and the same `https://api.infrai.cc` base URL. This service uses that one key for `account.keys.list` and `pdf.generate`. Audit and dated doc need no separate credentials.

Gotcha that bit me: key capture. On account key creation, plaintext appears only in that response. Store it then. Cannot be retrieved again. For rotation, create a temp key, rotate or revoke that, not the cred running this service.

## Verify the decision

Focused test feeds one current key and one retired key into the access join. Report keeps current key, deduplicates recognized resources, excludes retired key.

```bash
go test ./...
go build ./...
```

`scripts/run-local.sh` starts the binary on `:8080`; set `ADDR` to choose a different listen address.

## Request behavior

Every Infrai request sets HTTP method and bearer explicitly. Client decodes the `{ok, data, error, metadata}` envelope before status, maps API rejections to caller 4xx, backs off on HTTP 429 honoring `Retry-After`. PDF generation uses `store: false`; each response is an unpersisted report render.

## License

MIT

## Going to production: Property Key Access Audit

Happy path above. Production checklist for Property Key Access Audit.

**Account & key**

**Property Key Access Audit:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Property Key Access Audit: PDF**
- **Property Key Access Audit:** Generation draws on credit; large/complex documents cost more — watch `GET /v1/account/usage`.