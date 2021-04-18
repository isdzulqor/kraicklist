include .env
export

dev:
	@go run main.go api

seed:
	@env LOG_LEVEL=DEBUG go run main.go seed

# clean up all indexes docs on local
# for bleve index only
clean-index: 
	@rm -rf ./data/*.bleve
	@echo 'cleaning up docs on index..'

all: clean-index seed dev

# TODO: add readiness check
integration-test:
	@env PORT=7777 docker-compose -f docker-compose.test.yaml up $(build) -d
	@echo "wait for ES to be ready..."
	@sleep 25
	@-env PORT=7777 go test -v ./...
	@docker-compose -f docker-compose.test.yaml down
