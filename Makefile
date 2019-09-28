.PHONY: all format clean clean-all dep-init build-path delve-dbg-gate delev-dbg-svc

PROJECT_ROOT:=$(shell pwd)
export GOPATH:=$(PROJECT_ROOT)/build
export PATH:=$(PROJECT_ROOT)/bin:$(PATH)

BINARIES:=bin/wing

all: bin/dashboard format bin/wing 

format: build-path
	@for pkg in $$(cat "$(PROJECT_ROOT)/GOPACKAGES"); do \
		echo format "$$pkg"; 							\
		go fmt "$$pkg";									\
	done

bin/wing: dep-init bin/dashboard
	statik -src=$$(pwd)/bin/dashboard/
	@if [ "$${TYPE:=release}" = "debug" ]; then 						\
		go install -v -gcflags='all=-N -l' git.stuhome.com/Sunmxt/wing; \
	else																\
	    go install -v -ldflags='all=-s -w' git.stuhome.com/Sunmxt/wing; \
	fi;

bin/dashboard: build-path
	@if ! [ -z "$(SKIP_FE_BUILD)" ]; then 			\
		exit 0;										\
	fi; 											\
	cd $(PROJECT_ROOT)/dashboard;					\
	npm install; 									\
	npx webpack --mode=production; 					\
	cd ..;											\
	if [ ! -L "bin/dashboard" ]; then				\
		ln -s $$(pwd)/dashboard/dist bin/dashboard;	\
	fi


dep-init: build-path
	@cd $(PROJECT_ROOT); 																\
	if ! which statik > /dev/null && [ ! -e bin/statik ]; then 							\
		go get -u github.com/rakyll/statik;												\
	fi;																					\


delve-dbg-gate:
	@echo Not implemented.

delve-dbg-svc:
	@echo Not implemented.

# Common rules
build-path:
	@ENSURE_DIRS="bin build statik";				\
	for dir in $$ENSURE_DIRS; do				\
		if [ -e "$$dir" ]; then 				\
			if ! [ -d "$$dir" ]; then			\
				echo $$dir occupied.; 			\
				exit 1;							\
			fi;									\
		else									\
			mkdir $$dir;						\
		fi;										\
	done
	@if [ -e "build/bin" ]; then				\
		if ! [ -h "build/bin" ]; then			\
			echo build/bin occupied.;			\
		else									\
			rm build/bin;						\
		fi;										\
	fi
	@ln -s $$(pwd)/bin build/bin

clean:
	@if [ -e "build" ] && [ -d "build" ]; then \
		rm build -rf;							\
	fi

clean-all: clean
	@for binary in ${BINARIES}; do 	\
		if [ -e $$binary ]; then 	\
			rm $$binary;			\
			echo Remove $$binary; 	\
		fi;							\
	 done

