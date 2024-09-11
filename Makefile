build:
	go build -o server ./cmd/server

redis:
	cd deployments && docker compose up -d && echo 'Redis UI: http://localhost:9090'

instloadtest:
	go install fortio.org/fortio@latest

loadtesttoken:
	fortio load -t 20s -H 'API_KEY:abc123' http://localhost:8080

loadtestip:
	fortio load -t 20s http://localhost:8080

.PHONY: build redis instloadtest loadtestip loadtesttoken
