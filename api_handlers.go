package main

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"strconv"
	"math/rand"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db *gorm.DB
var rebateProgramCache *Cache

func reporting(c *gin.Context) {
	startDate, serr := time.Parse(dateFormat, 
		c.DefaultQuery("start_date", ""))
	endDate, eerr := time.Parse(dateFormat,
		c.DefaultQuery("end_date", ""))

	if serr != nil || eerr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"start_date and end_date required"})
		return
	}

	logger.Printf("start %v and end %v", startDate, endDate)  
	if result := db.Raw("SELECT SUM(claim_amount) FROM rebate_claims WHERE " +
	"claim_date > ? AND claim_date < ?", CustomTime{startDate}, CustomTime{endDate});
	result.Error != nil {
		logger.Fatalf("Error:%v", result.Error)
		c.JSON(500, gin.H{"error": "Failed to fetch claims"})
		return
	} else {
		var amount float64
		amount = 0
		rows,err := result.Rows()
		if err != nil {
			c.JSON(500, gin.H{"error":"Internal error"})
			return
		}
		rows.Next()
		rows.Scan(&amount)
		c.JSON(http.StatusOK, gin.H{"total_amount":amount})
		rows.Close()
		return
	}
}

func createRebateProgram(c *gin.Context) {
	var rebateProgram RebateProgram
	if err := c.ShouldBindJSON(&rebateProgram); err != nil {
		logger.Printf("Error in rebate unmarshalling:%v",err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if err := db.Create(&rebateProgram).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create rebate program"})
		return
	}

	c.JSON(200, gin.H{"message": "Rebate program created successfully",
		"id":rebateProgram.ID})
}

func getTransaction(id uint) (*Transaction, error) {
	transaction := &Transaction{}
	if err := db.First(&transaction, id).Error; err != nil {
		logger.Printf("Error:%v", err)
		return nil, err
	}
	return transaction, nil
}

//Fetch rebate helper, for use in api handlers
func getRebateProgram(id uint) (*RebateProgram, error) {
	var rebateProgram *RebateProgram
	if cachedRebateProgram, found := rebateProgramCache.Get(id); found {
		rebateProgram = cachedRebateProgram
	} else {
		// Cache miss - fetch from database and update cache
		rebateProgram = &RebateProgram{}
		if err := db.First(rebateProgram, id).Error; err != nil {
			logger.Printf("Error:In fetching rebate:%v", err)
			return nil, err
		}
		rebateProgramCache.Set(uint(id), *rebateProgram)
	}
	return rebateProgram, nil
}

func claimRebate(c *gin.Context) {
	var claim RebateClaim
	if err := c.ShouldBindJSON(&claim); err != nil {
		logger.Printf("Error:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Expected param:transaction_id"+
			"claim_amount,claim_date"})
		return
	}
	if claim.ClaimDate.Time.IsZero() {
		claim.ClaimDate.Time = time.Now()
	}
	
	//Query for creating new rebate claim
	query := `
	INSERT INTO rebate_claims (transaction_id, claim_amount, claim_status, claim_date)
	SELECT
		t.id AS transaction_id,
		CASE
			WHEN (t.transaction_date > r.start_date AND t.transaction_date < r.end_date)
			THEN (r.rebate_percentage * t.amount / 100)
			ELSE 0
		END AS claim_amount,
		'pending' AS claim_status,
		? AS claim_date
	FROM transactions t
	JOIN rebate_programs r ON t.rebate_program_id = r.id
	WHERE t.id = ?
	RETURNING id
	`
	var insertedIDs []uint
	if result := db.Raw(query, claim.ClaimDate, claim.TransactionID).Scan(&insertedIDs);
	result.Error != nil {
		logger.Fatalf("Error inserting claim %v", result.Error)
		c.JSON(500, gin.H{"error":"Failed to claim rebate"})
		return
	} else if len(insertedIDs) == 0 {
		c.JSON(500, gin.H{"error":"No claim added. Rebate or transaction not found"})
		return
	} else {
		logger.Printf("Inserted ids:%v", insertedIDs)
		c.JSON(http.StatusOK, gin.H{"message": "Claim added"})
		return
	}
}

// Submit Transaction with caching for Rebate Program
func submitTransaction(c *gin.Context) {
	var transaction Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check cache for rebate program
	var rebateProgram RebateProgram
	if cachedRebateProgram, found := rebateProgramCache.Get(transaction.RebateProgramID); found {
		rebateProgram = *cachedRebateProgram
	} else {
		// Cache miss - fetch from database and update cache
		if err := db.First(&rebateProgram, transaction.RebateProgramID).Error; err != nil {
			c.JSON(404, gin.H{"error": "Rebate Program not found"})
			return
		}
		rebateProgramCache.Set(transaction.RebateProgramID, rebateProgram)
	}

	// Set the current date for the transaction if not provided
	if transaction.TransactionDate.IsZero() {
		transaction.TransactionDate.Time = time.Now()
	}

	rebateValidityMessage := ""
	// Check if the transaction date is within the valid range of the rebate program
	if transaction.TransactionDate.Before(rebateProgram.StartDate.Time) ||
	transaction.TransactionDate.After(rebateProgram.EndDate.Time) {
		rebateValidityMessage = fmt.Sprintf(
			".Transaction date(%s) is out of rebate validity(%s - %s)",
			transaction.TransactionDate, rebateProgram.StartDate,
			rebateProgram.EndDate)
	}

	// Create the transaction record
	if err := db.Create(&transaction).Error; err != nil {
		c.JSON(500, gin.H{"error":"Failed to submit transaction"})
		return
	}

	c.JSON(200, gin.H{"message": "Transaction submitted successfully" +
		rebateValidityMessage, "id":transaction.ID})
}

// Fetch Rebate, Main function for /rebate/:id endpoint
func getRebate(c *gin.Context) {
	sRebateID := c.Param("rebate_id")
	rebateID, inputErr := strconv.Atoi(sRebateID)
	if inputErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Bad request. Expected integer in path param"})
		return
	}
	rebateProgram, rerr := getRebateProgram(uint(rebateID))
	if rerr != nil {
		c.JSON(404, gin.H{"error": "Rebate Program not found"})
		return
	}
	if responseJSON, err := json.Marshal(rebateProgram); err != nil {
		logger.Printf("Rebate program marshal error")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Unable to process the rebate program"})
		return
	} else {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, string(responseJSON))
	}
}

// Calculate Rebate with caching for Rebate Program
func calculateRebate(c *gin.Context) {
	transactionID := c.Param("transaction_id")
	var transaction Transaction
	if err := db.First(&transaction, transactionID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found"})
		return
	}

	// Check cache for rebate program
	var rebateProgram *RebateProgram
	if cachedRebateProgram, found := rebateProgramCache.Get(transaction.RebateProgramID); found {
		rebateProgram = cachedRebateProgram
	} else {
		// Cache miss - fetch from database and update cache
		rebateProgram = &RebateProgram{}
		if err := db.First(rebateProgram, transaction.RebateProgramID).Error; err != nil {
			c.JSON(404, gin.H{"error": "Rebate Program not found"})
			return
		}
		rebateProgramCache.Set(transaction.RebateProgramID, *rebateProgram)
	}

	if transaction.TransactionDate.Before(rebateProgram.StartDate.Time) ||
	transaction.TransactionDate.After(rebateProgram.EndDate.Time) {
		c.JSON(http.StatusOK, gin.H{"rebate_amount": 0,
			"message":fmt.Sprintf(
				"Transaction date(%s) is out of rebate validity(%s - %s)",
				transaction.TransactionDate, rebateProgram.StartDate,
				rebateProgram.EndDate)})
	} else {
		rebateAmount := transaction.Amount * rebateProgram.RebatePercentage / 100
		c.JSON(http.StatusOK, gin.H{"rebate_amount":rebateAmount})	
	}
}

// Simulate claim progress update (change some pending claims to approved or rejected)
func simulateProgress() {
	// Simulating random status change every 3 seconds
	for {
		time.Sleep(3 * time.Second)

		// Randomly pick a claim with pending status to update
		var claim RebateClaim
		if err := db.Where("claim_status = ?", "pending").Order("RANDOM()").First(&claim).Error; err == nil {
			// Simulate the change (either approve or reject)
			status := []string{"approved", "rejected"}
			newStatus := status[rand.Intn(2)]

			// Update claim status to simulate progress
			db.Model(&claim).Update("claim_status", newStatus)

			// Log for demo purposes
			logger.Printf("Claim ID %d status updated to %s", claim.ID, newStatus)
		}
	}
}

// Track progress of claims (approved vs pending)
func trackRebateClaimProgress(c *gin.Context) {
	// Using a single query to count the number of claims in each status (approved, pending, rejected)
	var progress ClaimProgress

	err := db.Table("rebate_claims").
		Select("SUM(CASE WHEN claim_status = 'approved' THEN 1 ELSE 0 END) AS approved_claims, "+
			"SUM(CASE WHEN claim_status = 'pending' THEN 1 ELSE 0 END) AS pending_claims, "+
			"SUM(CASE WHEN claim_status = 'rejected' THEN 1 ELSE 0 END) AS rejected_claims").
		Scan(&progress).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch rebate claim progress"})
		return
	}

	// Simulate progress (For demo purposes, let's randomly approve or reject some claims)
	go simulateProgress()

	// Return the current progress
	c.JSON(200, progress)
}

