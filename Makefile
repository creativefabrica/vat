test: .env.test
	go tool godotenv -f .env.test go tool gotestsum --format=testdox -- -cover ./...

test.coverage:
	mkdir -p ./coverage
	go tool godotenv -f .env.test go tool gotestsum --format=testdox -- -covermode=atomic -coverpkg=./... -coverprofile coverage/cover.out ./...

test.coverage.check: test.coverage
	go tool go-test-coverage --config=./.testcoverage.yml
	
test.coverage.treemap: test.coverage
	go tool go-cover-treemap -coverprofile coverage/cover.out > coverage.svg
	
gen:
	go generate ./...

lint.go:
	go tool golangci-lint run --fix

run.example:
	go tool godotenv -f .env.test go run ./example/main.go

.env.test:
	touch .env.test

deps.macos:
	brew bundle install

.PRECIOUS: .env.test 