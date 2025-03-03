test:
	go tool godotenv -f .env.test gotestsum --format=testdox -- -cover ./...

test.coverage:
	mkdir -p ./coverage
	go tool godotenv -f .env.test gotestsum --format=testdox -- -covermode=atomic -coverpkg=./... -coverprofile coverage/cover.out ./...

test.coverage.check: test.coverage
	go tool go-test-coverage --config=./.testcoverage.yml

gen:
	go generate ./...

lint.go:
	go tool golangci-lint run

modernize:
	go tool modernize ./...

vulcheck:
	go tool govulncheck ./...