package main

import (
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
	logger = log.Default()
}

func main() {
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
