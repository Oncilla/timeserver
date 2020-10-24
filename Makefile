gen:
	oapi-codegen -package gen -generate spec spec/timeserver.yml > pkg/api/gen/spec.gen.go
	oapi-codegen -package gen -generate types spec/timeserver.yml > pkg/api/gen/types.gen.go
	oapi-codegen -package gen -generate server spec/timeserver.yml > pkg/api/gen/server.gen.go

pkger:
	@echo "package slug" > slug.go
	pkger -o pkg/ui
	@rm slug.go

build: pkger
	go build -o bin/timeserver cmd/timeserver/*.go

run:
	go run cmd/timeserver/*.go
