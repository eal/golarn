

golarn: golarn.go
	go build -tags netgo -a -v -ldflags='-s' .
	#upx --ultra-brute golarn
	upx golarn

clean:
	rm -f golarn
