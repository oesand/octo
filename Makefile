
test:
	go test -race -v ./...

test_octogen:
	cd ./internal/octogen_tests && go run _run/main.go

autogen_test:
	cd ./internal/octogen_tests && go run _gen/main.go
