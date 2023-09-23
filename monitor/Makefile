all:
	git pull
	git checkout main
	GOOS=linux GOARCH=arm64 go build -o data/main_base main.go
	scp data/main_base my-aws:/home/ec2-user/main_base1

	git checkout linea
	GOOS=linux GOARCH=arm64 go build -o data/main_linea main.go
	scp data/main_linea my-aws:/home/ec2-user/main_linea1