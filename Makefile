.PHONY: all format clean start-dev-services stop-dev-services run-dev-worker run-dev-server force-run

PROJECT_ROOT:=$(shell pwd)
export GOPATH:=$(PROJECT_ROOT)/build
export PATH:=$(PROJECT_ROOT)/bin:$(PATH)

BINARIES:=bin/wing

all: format bin/wing 

format: statik
	@for pkg in $$(cat "$(PROJECT_ROOT)/GOPACKAGES"); do \
		echo format "$$pkg"; 							\
		go fmt "$$pkg";									\
	done

build:
	mkdir build

build/bin: bin
	ln -s $$(pwd)/bin build/bin

build/resource:
	mkdir build/resource

build/resource/sae: build/resource
	@[ ! -d "build/resource/sae" ] && mkdir -p build/resource/sae || true

bin:
	mkdir bin

statik:
	mkdir statik

build/resource/dashboard: build/resource
	mkdir build/resource/dashboard

dashboard/dist: build/resource/dashboard
	@[ ! -L "dashboard/dist" ] && ln -s $$(pwd)/build/resource/dashboard dashboard/dist; \
	if ! [ -z "$(SKIP_FE_BUILD)" ]; then 			\
		exit 0;										\
	fi; 											\
	cd $(PROJECT_ROOT)/dashboard;					\
	npm install; 									\
	npx webpack --mode=production; 					\
	cd ..

build/resource/sae/runtime: build/resource/sae
	tar -zcvf build/resource/sae/runtime -C controller/sae/runtime .

bin/statik: build/bin
	@[ ! -e "bin/statik" ] && go get -u github.com/rakyll/statik || true

bin/wing: statik build/resource/sae/runtime dashboard/dist bin/statik build/bin force-run
	@statik -src=$$(pwd)/build/resource/
	@if [ "$${TYPE:=release}" = "debug" ]; then 						\
		go install -v -gcflags='all=-N -l' git.stuhome.com/Sunmxt/wing; \
	else																\
	    go install -v -ldflags='all=-s -w' git.stuhome.com/Sunmxt/wing; \
	fi;

clean:
	[ -e "build" ] && [ -d "build" ] && rm build -rf

start-dev-services:
	docker-compose -f docker/compose-dev-service.yml up -d

stop-dev-services:
	docker-compose -f docker/compose-dev-service.yml down

run-dev-server: bin/wing
	bin/wing serve -config=./docker/compose-wing-dev-conf.yml -debug

run-dev-worker: bin/wing
	bin/wing worker -config=./docker/compose-wing-dev-conf.yml -debug

force-run:
	@true