

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

.PHONY: test
test: golarn.go golarn_test.go
	go test -cover -v

testloop:
	ag -l | entr make test


# Run golarn in dummy mode, for testing
.PHONY: dummy
dummy:
	GOLARN_TEMPLATE_PUSH='Push from {{.user_username}} on {{.project.name}}: {{if eq (print .total_commits_count) "1"}} {{- (index .commits 0).message| joinlines |truncatechars 50}} {{(index .commits 0).url}} {{else}} {{- .total_commits_count}} commits {{.project.web_url}}/compare/{{.before|slice 0 7}}...{{.after|slice 0 7}}{{end}}' \
	golarn -dummy
