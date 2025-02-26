run:
	go run main.go

build:
	docker build -t rcy0/linksheet:latest .

publish:
	docker push rcy0/linksheet:latest

deploy:
	fly deploy
