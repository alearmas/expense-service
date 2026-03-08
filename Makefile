.PHONY: build clean test tidy deploy

BINARY  := bootstrap
CMD     := ./cmd/main.go

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(BINARY) $(CMD)

# Target requerido por SAM (BuildMethod: makefile)
build-ExpenseFunction:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/$(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./... -v

tidy:
	go mod tidy

deploy: build
	sam build
	sam deploy
