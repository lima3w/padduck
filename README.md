
# IPAM Next

A modern IP Address Management (IPAM) platform built with Go, React, and PostgreSQL.

![CI Test Suite](https://gitea.lima3.dev/Lima3-Automations/ipam-next/actions/workflows/ci.yml/badge.svg)

## Purpose
Replace spreadsheet-based IP tracking with a structured, API-first system.

## Stack
- Backend: Go (Fiber)
- Frontend: React (Vite)
- Database: PostgreSQL
- Deployment: Docker Compose

## Run
docker compose up --build

## Deployment
Automated deployment to `gitea-runner.lab` is configured via `.gitea/workflows/deploy.yml`.
- Runs tests on all pushes to `main`
- Builds and deploys both backend and frontend services
- Verifies health endpoint before marking deployment successful

## Docs
- docs/architecture.md
- docs/roadmap.md
- docs/sprints.md
- docs/api.md
