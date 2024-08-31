package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite" 
)

var db *sqlx.DB
var mu sync.Mutex

type RawData struct {
	Timestamp string  `json:"timestamp"`
	KWhValue  float64 `json:"kWh_value"`
}

func main() {
	var err error
	db, err = sqlx.Open("sqlite", "data.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	// Set busy timeout (in milliseconds)
	_, err = db.Exec("PRAGMA busy_timeout = 5000") // 5 seconds
	if err != nil {
		log.Fatal("Error setting busy timeout:", err)
	}

	// Create tables if they don't exist
	createTables()

	err = importDataFromJSON("data.json")
	if err != nil {
		log.Fatal("Error importing data from JSON:", err)
	}

	// Start the background tasks
	go aggregateHourlyData()
	go aggregateDailyData()

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.GET("/api/hourly", getHourlyData)
	r.GET("/api/daily", getDailyData)
	r.Run(":8080")
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS raw_data (
		timestamp TEXT PRIMARY KEY,
		kWh_value REAL
	);
	CREATE TABLE IF NOT EXISTS hourly_aggregation (
		hourly_timestamp TEXT PRIMARY KEY,
		aggregated_kWh_value REAL
	);
	CREATE TABLE IF NOT EXISTS daily_aggregation (
		daily_timestamp TEXT PRIMARY KEY,
		aggregated_kWh_value REAL
	);
	`
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatal("Error creating tables:", err)
	}
}

func importDataFromJSON(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening JSON file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	var rawData []RawData
	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("error unmarshaling JSON data: %w", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	for _, item := range rawData {
		_, err := tx.Exec(`INSERT OR REPLACE INTO raw_data (timestamp, kWh_value) VALUES (?, ?)`,
			item.Timestamp, item.KWhValue)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error inserting data into raw_data: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func aggregateHourlyData() {
	for {
		now := time.Now().UTC()
		startOfHour := now.Truncate(time.Hour)
		endOfHour := startOfHour.Add(time.Hour)

		mu.Lock()
		tx, err := db.Beginx()
		if err != nil {
			log.Println("Error beginning transaction:", err)
			mu.Unlock()
			time.Sleep(time.Minute)
			continue
		}

		var total float64
		err = tx.Get(&total, `SELECT COALESCE(SUM(kWh_value), 0) FROM raw_data WHERE timestamp BETWEEN ? AND ?`, startOfHour.Format(time.RFC3339), endOfHour.Format(time.RFC3339))
		if err != nil {
			log.Println("Error fetching kWh data:", err)
			tx.Rollback()
			mu.Unlock()
			time.Sleep(time.Minute)
			continue
		}

		_, err = tx.Exec(`INSERT OR REPLACE INTO hourly_aggregation (hourly_timestamp, aggregated_kWh_value) VALUES (?, ?)`,
			startOfHour.Format(time.RFC3339), total)
		if err != nil {
			log.Println("Error inserting hourly aggregation:", err)
			tx.Rollback()
			mu.Unlock()
			continue
		}

		tx.Commit()
		mu.Unlock()
		time.Sleep(time.Hour)
	}
}

func aggregateDailyData() {
	for {
		now := time.Now().UTC()
		startOfDay := now.Truncate(24 * time.Hour)
		endOfDay := startOfDay.Add(24 * time.Hour)

		mu.Lock()
		tx, err := db.Beginx()
		if err != nil {
			log.Println("Error beginning transaction:", err)
			mu.Unlock()
			time.Sleep(time.Hour)
			continue
		}

		var total float64
		err = tx.Get(&total, `SELECT COALESCE(SUM(kWh_value), 0) FROM raw_data WHERE timestamp BETWEEN ? AND ?`, startOfDay.Format("2006-01-02T00:00:00Z"), endOfDay.Format("2006-01-02T00:00:00Z"))
		if err != nil {
			log.Println("Error fetching kWh data:", err)
			tx.Rollback()
			mu.Unlock()
			time.Sleep(time.Hour)
			continue
		}

		_, err = tx.Exec(`INSERT OR REPLACE INTO daily_aggregation (daily_timestamp, aggregated_kWh_value) VALUES (?, ?)`,
			startOfDay.Format("2006-01-02"), total)
		if err != nil {
			log.Println("Error inserting daily aggregation:", err)
			tx.Rollback()
			mu.Unlock()
			continue
		}

		tx.Commit()
		mu.Unlock()
		time.Sleep(24 * time.Hour)
	}
}

func getHourlyData(c *gin.Context) {
	var result []struct {
		HourlyTimestamp    string  `db:"hourly_timestamp"`
		AggregatedKWhValue float64 `db:"aggregated_kWh_value"`
	}
	err := db.Select(&result, "SELECT * FROM hourly_aggregation")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func getDailyData(c *gin.Context) {
	var result []struct {
		DailyTimestamp     string  `db:"daily_timestamp"`
		AggregatedKWhValue float64 `db:"aggregated_kWh_value"`
	}
	err := db.Select(&result, "SELECT * FROM daily_aggregation")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
