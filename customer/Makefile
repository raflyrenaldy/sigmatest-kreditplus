lint:
	golangci-lint run

migration:
	@read -p "migration file name:" module; \
	cd sigmatech/app/db/migrations && ~/go/bin/goose create $$module sql