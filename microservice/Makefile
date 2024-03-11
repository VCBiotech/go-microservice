ifeq ($(shell test -e './core/.envs/base.env' && echo -n yes),yes)
	include ./config/.env
endif

HELP_FUN = \
	%help; while(<>){push@{$$help{$$2//'options'}},[$$1,$$3] \
	if/^([\w-_]+)\s*:.*\#\#(?:@(\w+))?\s(.*)$$/}; \
    print"$$_:\n", map"  $$_->[0]".(" "x(20-length($$_->[0])))."$$_->[1]\n",\
    @{$$help{$$_}},"\n" for keys %help; \

args = $(filter-out $@,$(MAKECMDGOALS))


# Commands
help: ##@Help Show this help
	@echo -e "Usage: make [target] ...\n"
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

dev: ##@Application Run development server with hot reload
	lsof -t -i tcp:3000 | xargs kill -9
	air

build: ##@Build build the application using Go compiler
	docker build --tag 'dev' .

run:  ##@Application Run application locally
	docker run --detach 'dev'

shell:  #@Application Run Poetry Shell
	echo "To do!"

migrate:  ##@Application Apply migrations
	migrate --help

new-migration:  ##@Application Create migrations
	migrate create -ext sql -dir migrations -seq $(args)

.PHONY: test
test:  ##@Testing Test application with Django Testing Library
	go test -v ./test

lint: ##@Linting the application with ruff
	echo "To do!"

format: ##@Formatting the application with ruff
	echo "To do!"

type-check: ##@Type checking the application with mypy
	echo "To do!"

clean: ##@Utility Clean the repository from build artifacts
	rm bin/microservice

%::
	echo $(MESSAGE)
