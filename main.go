package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Input records
type Record struct {
	Timestamp string  `json:"timestamp"`
	KWhValue  float64 `json:"kWh_value"`
}

// Output CSV
type HourlyData struct {
	Hour     string  `json:"hour"`
	KWhValue float64 `json:"kWh_value"`
}

type DailyData struct {
	Date     string  `json:"date"`
	KWhValue float64 `json:"kWh_value"`
}

func parseTimestamp(ts string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation(time.RFC3339, ts, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func aggregateData(records []Record, loc *time.Location) (map[string]map[string]float64, map[string]float64) {
	hourlyData := make(map[string]map[string]float64)
	dailyData := make(map[string]float64)

	for _, record := range records {
		t, err := parseTimestamp(record.Timestamp, loc)
		if err != nil {
			fmt.Println("Error parsing timestamp:", err)
			continue
		}

		month, day := int(t.Month()), t.Day()
		hour := fmt.Sprintf("%02d:00:00", t.Hour())
		date := fmt.Sprintf("%02d/%02d", day, month)

		if _, ok := hourlyData[date]; !ok {
			hourlyData[date] = make(map[string]float64)
		}
		hourlyData[date][hour] += record.KWhValue

		dailyData[date] += record.KWhValue
	}

	return hourlyData, dailyData
}

func saveCSV(filePath string, data interface{}, headers []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return err
	}

	switch v := data.(type) {
	case map[string]map[string]float64:
		for date, hours := range v {
			for hour, value := range hours {
				record := []string{date, hour, fmt.Sprintf("%.2f", value)}
				if err := writer.Write(record); err != nil {
					return err
				}
			}
		}
	case map[string]float64:
		for date, value := range v {
			record := []string{date, fmt.Sprintf("%.2f", value)}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}
	return nil
}

func saveHourlyCSV(data map[string]map[string]float64) {
	for date, hours := range data {
		filePath := fmt.Sprintf("output_data/hourly_data_%s.csv", strings.ReplaceAll(date, "/", "_"))
		if err := saveCSV(filePath, hours, []string{"Date", "Hour", "kWh Value"}); err != nil {
			fmt.Println("Error saving hourly CSV:", err)
		}
	}
}

func saveDailyCSV(data map[string]float64) {
	filePath := "output_data/daily_data.csv"
	if err := saveCSV(filePath, data, []string{"Date", "kWh Value"}); err != nil {
		fmt.Println("Error saving daily CSV:", err)
	}
}

func main() {
	file, err := os.Open("data.json")
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer file.Close()

	var records []Record
	if err := json.NewDecoder(file).Decode(&records); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// Use UTC as the default time zone
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		fmt.Println("Error loading time zone:", err)
		return
	}

	hourlyData, dailyData := aggregateData(records, loc)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("output_data", os.ModePerm); err != nil {
		fmt.Println("Error creating output directory:", err)
		return
	}

	// Save data sequentially
	saveHourlyCSV(hourlyData)
	saveDailyCSV(dailyData)

	fmt.Println("Data saved successfully.")
}
