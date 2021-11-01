build:
	go mod tidy
	go mod download
	go build -o ./sops-precommit ./main.go
