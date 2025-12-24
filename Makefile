BINARY=rttlog
CMD=./cmd/rttlog
DIST=dist

.PHONY: fmt vet test race integration build dist clean

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./... -count=1

race:
	go test ./... -race -count=1

integration:
	RTTLOG_INTEGRATION=1 go test -tags=integration ./... -count=1

build:
	go build -o $(BINARY) $(CMD)

dist: clean
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o $(DIST)/$(BINARY)-linux-amd64 $(CMD)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -o $(DIST)/$(BINARY)-linux-arm64 $(CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(BINARY)-windows-amd64.exe $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o $(DIST)/$(BINARY)-darwin-amd64 $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o $(DIST)/$(BINARY)-darwin-arm64 $(CMD)

clean:
	rm -rf $(DIST)
