run:
	air

build:
	docker build -t rcy0/linksheet:latest .

publish:
	docker push rcy0/linksheet:latest

deploy:
	fly deploy
