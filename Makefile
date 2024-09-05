.PHONY: install
install:
	go build
	cp -f aws-prompt ~/.local/bin
