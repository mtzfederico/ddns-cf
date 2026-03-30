# https://tutorialedge.net/golang/makefiles-for-go-developers/
# https://stackoverflow.com/questions/20829155/how-to-cross-compile-from-windows-to-linux

build:
	go mod download
	go generate
	go build -o bin/ddns-cf

update:
	cp /home/fedemtz/ddns-cf/bin/ddns-cf /home/fedemtz/ddns-cf/bin/prev-ddns-cf-$(shell date +%Y-%m-%d_%H-%M-%S)
	git pull
	$(MAKE) build

dev:
	go generate
	go build -o bin/ddns-cf-dev
