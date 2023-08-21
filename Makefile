.PHONY: readme test

test: ## runs the tests and generates a coverage report
	@ go test -count=1 -v ./... -covermode=count -coverprofile=coverage.out
	@ go tool cover -func=coverage.out -o=coverage.out

lint: ## runs the linters and makes sure the README.md is up to date
	@ golangci-lint run 
	@ bash -c "cmp -s README.md <(goreadme -credit=false -skip-sub-packages)" || (echo "README.md is out of date. Please run 'make readme'" && exit 1)

readme: ## generates the readme file for the package from the godocs
	@ goreadme -credit=false -skip-sub-packages > README.md