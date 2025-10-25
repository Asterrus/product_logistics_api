.PHONY: build
build:
	go build cmd/logistic-product-api/main.go

.PHONY: test
test_r:
	go test -v ./internal/app/retranslator/
test_w:
	go test -v ./internal/app/workerpool/
