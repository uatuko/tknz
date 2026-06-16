binary = .build/idaas

pbdir     = pb
pbdir-int = internal/pb
pbdir-klara = internal/klarapb

protodir  = proto
sourcedir = .

sources    := $(shell find $(sourcedir) -type f -name '*.go')
protos     := $(shell find $(protodir) -type f -name '*.proto' -not -path '*/detail/*' -not -path '*/klara/*')
protos-int := $(shell find $(protodir) -type f -name '*.proto' -path '*/detail/*')
protos-klara := $(shell find $(protodir) -type f -name '*.proto' -path '*/klara/*')

githubrepo = $(shell awk 'NR==1 {print $$2; exit}' go.mod)

%:
	@:

$(binary): $(sources)
	go build -o $(binary) main.go

all: $(binary)

clean:
	go clean
	if [ -f $(binary) ] ; then rm $(binary); fi

lint:
	gofmt -s -d -e $(sourcedir)
	npm run lint
ifeq (, $(shell which clang-format))
	echo '\033[1;41m WARN \033[0m clang-format not found, not linting files';
else
	clang-format --style=file --dry-run $(protos)
endif

lint\:ci:
	test -z "$(shell gofmt -s -l $(sourcedir))"
	npm run lint
	clang-format --style=file --dry-run --Werror $(protos-int) $(protos)

lint\:fix:
	gofmt -s -l -w $(sourcedir)
	npm run lint:fix
	clang-format --style=file -i $(protos-int) $(protos)

protoc-gen-go:
ifeq (, $(shell which protoc-gen-go))
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
endif

protoc-gen-go-grpc:
ifeq (, $(shell which protoc-gen-go-grpc))
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
endif

protoc-int: $(protos-int) protoc-gen-go protoc-gen-go-grpc
	protoc \
		--go_out=$(pbdir-int) --go_opt=module=$(githubrepo)/$(pbdir-int) \
		--go-grpc_out=$(pbdir-int) --go-grpc_opt=module=$(githubrepo)/$(pbdir-int) \
		$(protos-int)

protoc-klara: $(protos-klara) protoc-gen-go protoc-gen-go-grpc
	protoc \
		--go_out=$(pbdir-klara) --go_opt=module=$(githubrepo)/$(pbdir-klara) \
		--go-grpc_out=$(pbdir-klara) --go-grpc_opt=module=$(githubrepo)/$(pbdir-klara) \
		$(protos-klara)

protoc: protoc-int protoc-klara
	protoc \
		--go_out=$(pbdir) --go_opt=module=$(githubrepo)/$(pbdir) \
		--go-grpc_out=$(pbdir) --go-grpc_opt=module=$(githubrepo)/$(pbdir) \
		$(protos)

run:
	npm run build
	go run -race . -debug \
		-addr localhost:8080 \
		-mail-addr localhost:50151 \
		$(filter-out $@,$(MAKECMDGOALS))

test:
	go test -race -tags=test ./...

test\:coverage:
	go test -coverprofile=.coverage.out -covermode=atomic -race -tags=test ./...
