test: .env.test
	go tool godotenv -f .env.test go tool gotestsum --format=testdox -- -cover ./...

test.coverage:
	mkdir -p ./coverage
	go tool godotenv -f .env.test go tool gotestsum --format=testdox -- -covermode=atomic -coverpkg=./... -coverprofile coverage/cover.out ./...

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

run.example:
	go tool godotenv -f .env.test go run ./example/main.go

.env.test:
	touch .env.test

.PRECIOUS: .env.test