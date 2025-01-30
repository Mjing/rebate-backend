package main
import (
	"flag"
	apis "rebate-backend/src/restapis"
)

func main() {
	port := flag.String("port", ":8080", "Host address to run server on")
	logFile := flag.String("log", "", "Log file name. default:stdout")
	flag.Parse()
	if (*logFile != "") {
		apis.SetLogFile(*logFile)
	}
	apis.RunServer(*port)
}
