package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

// RebateProgram model
type RebateProgram struct {
	ID                uint      `gorm:"primaryKey"`
	ProgramName       string    `json:"program_name" gorm:"not null"`
	RebatePercentage  float64   `json:"rebate_percentage" gorm:"not null"`
	StartDate         CustomTime`json:"start_date" gorm:"not null"`
	EndDate           CustomTime`json:"end_date" gorm:"not null"`
	EligibilityCriteria string  `json:"eligibility_criteria" gorm:"not null"`
}

// Transaction model
type Transaction struct {
	ID              uint		`gorm:"primaryKey" json:"id"`
	Amount          float64		`gorm:"not null" json:"amount"`
	TransactionDate CustomTime	`gorm:"not null" json:"transaction_date"`
	RebateProgramID uint		`gorm:"not null;index" json:"rebate_program_id"`
}

// ClaimStatus type
type ClaimStatus string

// RebateClaim model
type RebateClaim struct {
	ID            uint        `gorm:"primaryKey;" json:"id"`
	TransactionID uint        `gorm:"not null;" json:"transaction_id"`
	ClaimAmount   float64     `gorm:"not null;" json:"claim_amount"`
	ClaimStatus   ClaimStatus `json:"claim_status" gorm:"type:TEXT;check:claim_status IN ('pending', 'approved', 'rejected')"`
	ClaimDate     CustomTime  `gorm:"not null" json:"claim_date"`
}

const (
	Pending  ClaimStatus = "pending"
	Approved ClaimStatus = "approved"
	Rejected ClaimStatus = "rejected"
)

// BeforeCreate hook to validate the ClaimStatus before inserting into the database
func (claim *RebateClaim) BeforeCreate(tx *gorm.DB) (err error) {
	// Ensure that ClaimStatus is valid
	if claim.ClaimStatus != Pending && claim.ClaimStatus != Approved && claim.ClaimStatus != Rejected {
		return fmt.Errorf("invalid claim status: %s", claim.ClaimStatus)
	}
	return nil
}

type ClaimProgress struct {
	ApprovedClaims int64 `json:"approved_claims"`
	PendingClaims  int64 `json:"pending_claims"`
	RejectedClaims int64 `json:"rejected_claims"`
}
