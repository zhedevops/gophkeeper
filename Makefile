.PHONY: server
server:
	go run ./cmd/server/main.go

.PHONY: register
register:
	go run . register -l $(login)

.PHONY: login
login:
	go run . login -l $(login)

.PHONY: create-credentials
create-credentials:
	go run . vault create -t $(type) -m "$(meta)" -l $(login) -p $(password)

.PHONY: create-text
create-text:
	go run . vault create -t $(type) -m "$(meta)" -d "$(data)"

.PHONY: create-binary-file
create-binary-file:
	go run . vault create -t $(type) -m "$(meta)" -f $(file)

.PHONY: create-binary-data
create-binary-data:
	go run . vault create -t $(type) -m "$(meta)" -d $(data)

.PHONY: create-card
create-card:
	go run . vault create -t $(type) -m "$(meta)" -n $(number) -o "$(holder)" -e $(expiry) -c $(cvv)

.PHONY: get
get:
	go run . vault get -i $(id)

.PHONY: list
list:
	go run . vault list

.PHONY: delete
delete:
	go run . vault delete -i $(id)

.PHONY: certs
certs:
	mkdir -p certs
	openssl req \
		-x509 \
		-newkey rsa:4096 \
		-sha256 \
		-days 365 \
		-nodes \
		-keyout certs/server.key \
		-out certs/server.crt \
		-config certs/openssl.cnf

.PHONY: certs-tests
certs-tests:
	mkdir -p testdata
	openssl req \
		-x509 \
		-newkey rsa:4096 \
		-sha256 \
		-days 365 \
		-nodes \
		-keyout testdata/server.key \
		-out testdata/server.crt \
		-config certs/openssl.cnf

gmod:
	go mod tidy

test:
	go test -count 1 ./...

testcov:
	go test -count 1 ./... -cover

vet:
	go vet ./...

vetstat:
	go vet -vettool=$(which statictest) ./...