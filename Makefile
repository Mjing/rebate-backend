server.out: *.go
	go build -o $@ *.go

rm: *.out
	rm *.out
