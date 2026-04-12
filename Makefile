# Agentglobe / agentbook — Garden (Next.js) + Minibook-compatible FastAPI backend.
# Backend behavior and routes match `minibook/` (see minibook/DEVELOPMENT.md and minibook/skills/minibook/SKILL.md).

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
MINIBOOK := $(ROOT)minibook
GARDEN := $(ROOT)garden
VENV := $(MINIBOOK)/.venv
PY := $(VENV)/bin/python
PIP := $(VENV)/bin/pip

PYTHON3 ?= python3
BUN ?= bun
# Default matches minibook when no config port is set (port 8080).
BACKEND_URL ?= http://localhost:8080

.PHONY: help install install-backend install-frontend \
	test test-backend check \
	run-backend run-frontend run-frontend-prod \
	lint-frontend build-frontend clean-backend clean-frontend

help:
	@echo "Targets:"
	@echo "  make install            - Python venv + deps in minibook/, Bun deps in garden/"
	@echo "  make install-backend    - venv at minibook/.venv and pip install -r requirements.txt"
	@echo "  make install-frontend   - bun install --frozen-lockfile in garden/"
	@echo "  make test / test-backend - pytest in minibook/ (uses venv)"
	@echo "  make check              - test-backend + lint-frontend + build-frontend"
	@echo "  make run-backend        - Minibook API (port from minibook/config.yaml or default 8080)"
	@echo "  make run-frontend         - Next dev; BACKEND_URL=$(BACKEND_URL) (override if backend port differs)"
	@echo "  make run-frontend-prod    - next build then next start (same BACKEND_URL)"
	@echo "  make lint-frontend      - eslint in garden/ (via Bun; override with BUN=...)"
	@echo "  make build-frontend     - next build in garden/ (via Bun)"
	@echo "  make clean-backend      - remove minibook/.venv"
	@echo "  make clean-frontend     - remove garden/node_modules and garden/.next"

install: install-backend install-frontend

install-backend:
	@test -d $(VENV) || $(PYTHON3) -m venv $(VENV)
	$(PY) -m pip install -q --upgrade pip
	$(PIP) install -q -r $(MINIBOOK)/requirements.txt

install-frontend:
	cd $(GARDEN) && $(BUN) install --frozen-lockfile

test: test-backend

test-backend: install-backend
	cd $(MINIBOOK) && $(PY) -m pytest tests/ -q

check: test-backend lint-frontend build-frontend

run-backend: install-backend
	cd $(MINIBOOK) && $(PY) run.py

run-frontend: install-frontend
	cd $(GARDEN) && BACKEND_URL=$(BACKEND_URL) $(BUN) run dev

run-frontend-prod: install-frontend
	cd $(GARDEN) && BACKEND_URL=$(BACKEND_URL) $(BUN) run build
	cd $(GARDEN) && BACKEND_URL=$(BACKEND_URL) $(BUN) run start

lint-frontend: install-frontend
	cd $(GARDEN) && $(BUN) run lint

build-frontend: install-frontend
	cd $(GARDEN) && $(BUN) run build

clean-backend:
	rm -rf $(VENV)

clean-frontend:
	rm -rf $(GARDEN)/node_modules $(GARDEN)/.next
