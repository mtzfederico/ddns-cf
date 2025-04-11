# https://tutorialedge.net/golang/makefiles-for-go-developers/
# https://stackoverflow.com/questions/20829155/how-to-cross-compile-from-windows-to-linux

build:
	go get .
	go build -o bin/ddns-cf

update:
	git pull
	make build

dev:
	go build -o bin/ddns-cf-dev
