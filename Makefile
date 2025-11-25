## Makefile - convenience targets for running the sample commands

.PHONY: backup restore

backup:
	@echo "Running backup..."
	@go run ./cmd/backup

build:
	@echo "Build backup and restore program..."
	go build -o tmp/backup ./cmd/backup 
	go build -o tmp/restore ./cmd/restore

restore:
	@echo "Running restore..."
	@go run ./cmd/restore

rclone:
	@echo "(TBD) Start RClone..."
