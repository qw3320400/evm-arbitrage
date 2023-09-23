chain := $(shell git symbolic-ref -q HEAD | cut -d"/" -f 3)

build:
	echo $(chain)
	git pull
	git checkout $(chain)
	GOOS=linux GOARCH=arm64 go build -o data/main_$(chain) main.go
	scp data/main_$(chain) my-aws:/home/ec2-user/main_$(chain)1
	ssh my-aws 'mv -f main_$(chain)1 main_$(chain) && sudo supervisorctl restart main_$(chain)'