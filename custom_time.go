package main
import (
	"fmt"
	"time"
	"encoding/json"
	"database/sql/driver"
)

type CustomTime struct {
	time.Time
}
//Function for sql drivers
func (ct *CustomTime) Scan(value interface{}) error {
	date, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("Failed to read date data.")
	}
	ct.Time = date
	return nil
}

func (ct CustomTime) Value() (driver.Value, error) {
	return ct.Time, nil
}

const dateFormat = "02/01/2006"
func (ct CustomTime) String() string {
	return ct.Format(dateFormat)
}

// MarshalJSON customizes the JSON encoding for CustomTime
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	// Format the time in "yyyy-mm-dd" format
	return json.Marshal(ct.String())
}

// UnmarshalJSON customizes the JSON decoding for CustomTime
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// Parse the time in "yyyy-mm-dd" format
	str := string(data)
	str = str[1:len(str)-1]
	parsedTime, err := time.Parse(dateFormat, str)
	if err != nil {
		return err
	}
	ct.Time = parsedTime
	return nil
}
