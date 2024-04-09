build:
	go build -o pulse-api .
.PHONY: local
local: build
	./pulse-api