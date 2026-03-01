default: test

.PHONY: test
test:
	go test -v ./... -count=1

.PHONY: gen
gen:
	go generate ./...
