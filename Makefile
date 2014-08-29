
CODEZ=$(shell find src -name '*.go')
CODEZ+=$(shell find $(GOPATH)/src/github.com/porty/emitter -name '*.go')

all: bin bin/ecat-osx bin/ecat-linux bin/ecat-pi

bin:
	mkdir bin

bin/ecat-osx: $(CODEZ)
	cd src && GOOS=darwin go build -o ../$@

bin/ecat-linux: $(CODEZ)
	cd src && GOOS=linux go build -o ../$@

bin/ecat-pi: $(CODEZ)
	cd src && GOARCH=arm GOOS=linux GOARM=5 go build -o ../$@

test:
	cd src && go test ./...

clean:
	rm -rf bin

ssh-gateway:
	ssh 192.168.73.129

ssh-client:
	ssh 192.168.178.128

upload: bin/ecat-linux
	scp bin/ecat-linux 192.168.178.128:/home/shorty/

.PHONY: test ssh-gateway ssh-client upload clean
