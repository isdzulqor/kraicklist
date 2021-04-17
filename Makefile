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
