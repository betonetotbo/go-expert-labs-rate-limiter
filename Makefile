build:
	go build -o server ./cmd/server

instloadtest:
	go install fortio.org/fortio@latest

loadtest:
	fortio load -t 20s http://localhost:3000

.PHONY: build instloadtest loadtest
