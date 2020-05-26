.PHONY: deps build update-scripts

deps:
	go get -u .

build:
	go build main.go
	mv main ~/dyson-controller

update-scripts:
	cp scripts/* ~/
