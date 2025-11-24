## Makefile - convenience targets for running the sample commands

.PHONY: backup restore

backup:
	@echo "Running backup..."
	@go run ./cmd/backup

restore:
	@echo "Running restore..."
	@go run ./cmd/restore

rclone:
	@echo "(TBD) Start RClone..."
