# pgview demo kit — Makefile convenience wrapper
# Requires: make  (optional — use start.sh directly if you don't have it)
# The primary entry point is start.sh; this wraps it for users who prefer make targets.

COMPOSE := $(shell command -v docker 2>/dev/null && echo "docker compose" || \
                   command -v podman 2>/dev/null && echo "podman compose")

.PHONY: try stop clean

## try: start the demo database, download pgview, and launch
try:
	bash start.sh

## stop: tear down the demo database container
stop:
	$(COMPOSE) down

## clean: remove the downloaded pgview binary
clean:
	rm -f pgview
