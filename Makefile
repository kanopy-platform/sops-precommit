build:
	go mod tidy
	go mod download
	go install ./cmd/sops-precommit/
