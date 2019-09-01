PROJECT_NAME = 'sensorctl'

# Version
VERSION = `date +%y.%m`

# If unable to grab the version, default to N/A
ifndef VERSION
    VERSION = "n/a"
endif

#
# Makefile options
#


# State the "phony" targets
.PHONY: all clean build install uninstall

all: build

build: clean
	@echo 'Building ${PROJECT_NAME}...'
	@go build -ldflags '-s -w -X main.version='${VERSION}

clean:
	@echo 'Cleaning...'
	@go clean

install: build
	@echo Installing executable file to /usr/local/bin/${PROJECT_NAME}
	@sudo cp ${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}
	@sudo chmod +x /usr/local/bin/${PROJECT_NAME}

uninstall: clean
	@echo Removing executable file from /usr/local/bin/${PROJECT_NAME}
	@sudo rm -f /usr/local/bin/${PROJECT_NAME}
