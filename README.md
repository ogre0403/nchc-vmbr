# cloud-sdk-sample

This sample demonstrates using the `Zillaforge` Cloud SDK to create VM snapshots and export them.

Refactor notes:
- The backup logic is now implemented in `internal/backup` with `LoadConfigFromEnv` and `Run` functions.
- `cmd/backup` and `cmd/recover` act as thin CLIs that delegate to the internal packages.

How to run:
1. Create a `.env` file with required variables: API_PROTOCOL, API_HOST, API_TOKEN, PROJECT_SYS_CODE, SRC_VM, SNAPSHOT_NAME, CS_BUCKET

Optional:
- `DATE_TAG_FORMAT` - layout used to format the date tag for snapshots. If omitted, the default format `2006-01-02-15-04` is used. Use Go time layout strings for the format.
2. Execute the backup command:

```bash
go run ./cmd/backup
```

This repo uses `godotenv` to load variables and has a test for the `internal/backup` config loader.

Recover flow:
- Required env vars: API_PROTOCOL, API_HOST, API_TOKEN, PROJECT_SYS_CODE, RESTORE_IMAGE_NAME, RESTORE_IMAGE_PATH or (RESTORE_CS_BUCKET + DATE_TAG), RESTORE_FLAVOR_ID, RESTORE_NETWORK_ID
- Optional: RESTORE_SG_ID, RESTORE_VM_PREFIX, RESTORE_PASSWORD_BASE64 (or RESTORE_PASSWORD), RESTORE_OS_TYPE

To execute recover:
```bash
go run ./cmd/recover
```
