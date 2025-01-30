package restapis

import (
	"os"
	"log"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func initDB() {
	var err error
	db, err = gorm.Open("sqlite3", "./rebate.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&RebateProgram{}, &Transaction{}, &RebateClaim{})
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_rebate_claims_id_transaction_id"+
		" ON rebate_claims (id, transaction_id);")
}

var logger *log.Logger
func initMain() {
	initDB()
	rebateProgramCache = NewCache(10 * time.Minute)
	if logger == nil {
		logger = log.Default()
	}
}

func SetLogFile(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v\n", err)
		file.Close()
	}
	logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func RunServer(hostaddr string) {
	initMain()
	defer db.Close()

	r := gin.Default()

	// Define API routes
	r.POST("/rebate_program", createRebateProgram)
	r.POST("/transaction", submitTransaction)
	r.GET("/calculate_rebate/:transaction_id", calculateRebate)
	r.GET("/rebate_program/:rebate_id", getRebate)
	r.POST("/claim_rebate", claimRebate)
	r.GET("/reporting", reporting)
	r.GET("/rebate_claims/progress", trackRebateClaimProgress)

	// Start the server
	r.Run(":8080")
}
