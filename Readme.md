package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type DataEntry struct {
	Timestamp string  `json:"timestamp"`
	KWhValue  float64 `json:"kWh_value"`
	SourceID  string  `json:"source_id,omitempty"`
}

func parseData(input []byte) ([]DataEntry, error) {
	var data []DataEntry
	err := json.Unmarshal(input, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func aggregateHourly(data []DataEntry) map[string]map[string]float64 {
	result := make(map[string]map[string]float64)
	for _, entry := range data {
		timestamp, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			continue
		}
		dateKey := timestamp.Format("2006-01-02")
		hourKey := timestamp.Format("15:00:00")

		if _, exists := result[dateKey]; !exists {
			result[dateKey] = make(map[string]float64)
		}
		result[dateKey][hourKey] += entry.KWhValue
	}
	return result
}

func aggregateDaily(data []DataEntry) map[string]float64 {
	result := make(map[string]float64)
	for _, entry := range data {
		timestamp, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			continue
		}
		dateKey := timestamp.Format("02/01")
		result[dateKey] += entry.KWhValue
	}
	return result
}

func printHourlyTables(data map[string]map[string]float64) {
	for date, hourlyData := range data {
		fmt.Printf("Hourly Data for %s\n", date)
		fmt.Println(strings.Repeat("-", 40))

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
		fmt.Fprintln(writer, "Hour\tkWh Value")

		for hour, value := range hourlyData {
			fmt.Fprintf(writer, "%s\t%.2f\n", hour, value)
		}

		writer.Flush()
		fmt.Println()
	}
}

func printDailyTable(data map[string]float64) {
	fmt.Println("Daily Data")
	fmt.Println(strings.Repeat("-", 20))

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(writer, "Date\tkWh Value")

	for date, value := range data {
		fmt.Fprintf(writer, "%s\t%.2f\n", date, value)
	}

	writer.Flush()
	fmt.Println()
}

func saveCSV(filename string, headers []string, rows [][]string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating file %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		log.Fatalf("Error writing headers to file %s: %v", filename, err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			log.Fatalf("Error writing row to file %s: %v", filename, err)
		}
	}
}

func saveHourlyCSV(directory string, data map[string]map[string]float64) {
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory %s: %v", directory, err)
	}

	var wg sync.WaitGroup
	for date, hourlyData := range data {
		wg.Add(1)
		go func(date string, hourlyData map[string]float64) {
			defer wg.Done()
			filename := filepath.Join(directory, fmt.Sprintf("hourly_data_%s.csv", date))
			headers := []string{"Hour", "kWh Value"}

			var rows [][]string
			for hour, value := range hourlyData {
				rows = append(rows, []string{hour, fmt.Sprintf("%.2f", value)})
			}

			saveCSV(filename, headers, rows)
		}(date, hourlyData)
	}

	wg.Wait()
}

func saveDailyCSV(directory string, data map[string]float64) {
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory %s: %v", directory, err)
	}

	filename := filepath.Join(directory, "daily_data.csv")
	headers := []string{"Date", "kWh Value"}

	var rows [][]string
	for date, value := range data {
		rows = append(rows, []string{date, fmt.Sprintf("%.2f", value)})
	}

	saveCSV(filename, headers, rows)
}

func main() {
	inputFile := "data.json"
	file, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	data, err := parseData(file)
	if err != nil {
		log.Fatalf("Error parsing input data: %v", err)
	}

	hourlyData := aggregateHourly(data)
	dailyData := aggregateDaily(data)

	printHourlyTables(hourlyData)
	printDailyTable(dailyData)

	directory := "output_data"

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		saveHourlyCSV(directory, hourlyData)
	}()

	go func() {
		defer wg.Done()
		saveDailyCSV(directory, dailyData)
	}()

	wg.Wait()
}
