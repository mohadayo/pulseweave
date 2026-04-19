.PHONY: test test-python test-go test-ts lint up down build clean

test: test-python test-go test-ts

test-python:
	cd services/auth-service && pip install -r requirements.txt -q && pytest -v

test-go:
	cd services/metrics-engine && go test -v ./...

test-ts:
	cd services/dashboard-api && npm install --silent && npm test

lint: lint-python lint-go lint-ts

lint-python:
	cd services/auth-service && flake8 app.py test_app.py --max-line-length=120

lint-go:
	cd services/metrics-engine && go vet ./...

lint-ts:
	cd services/dashboard-api && npx eslint src/ --ext .ts

up:
	docker compose up -d --build

down:
	docker compose down

build:
	docker compose build

clean:
	docker compose down -v --rmi local
