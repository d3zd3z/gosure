# Build as desired:

all: gosure

gosure: .force
	GOPATH=$(PWD) go build gosure

clean:
	rm -f bin/*
	rm -f src/*/*.6
	rm -rf src/*/_obj
	rm -f src/*/pr[0-9]
	rm -f src/*/pr[0-9][0-9]
	rm -f src/*/pr[0-9][0-9][0-9]

.PHONY: .force
.force:
