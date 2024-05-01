run: 
	go run github.com/cosmtrek/air 

build:
	go build -o ./bin/app .

.PHONY:
	run