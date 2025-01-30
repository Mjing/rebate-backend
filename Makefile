server.out: cmd/main.go restapi.obj
	go build -o $@ cmd/main.go

restapi.obj: src/restapis/*.go
	go build -o $@ src/restapis/*.go

rm:
	rm *.out *.obj
