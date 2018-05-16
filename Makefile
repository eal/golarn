

golarn: golarn.go
	go build -tags netgo -a -v -ldflags='-s' .
	#upx --ultra-brute golarn
	upx golarn

clean:
	rm -f golarn

docker: golarn
	sudo docker build -t golarn:latest .
	sudo docker tag golarn:latest docker.io/datatyp/golarn:latest

push:
	sudo docker push docker.io/datatyp/golarn:latest
