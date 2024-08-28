package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

// Input records
type Record struct {
	Timestamp string  `json:"timestamp"`
	KWhValue  float64 `json:"kWh_value"`
}

// Output CSV
type HourlyData struct {
	Date     string  `json:"date"`
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
		hour := fmt.Sprintf("%02d:00", t.Hour())
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
		// Convert data to a slice for sorting
		var records []HourlyData
		for date, hours := range v {
			for hour, value := range hours {
				records = append(records, HourlyData{
					Date:     date,
					Hour:     hour,
					KWhValue: value,
				})
			}
		}
		// Sort records by date and hour
		sort.Slice(records, func(i, j int) bool {
			if records[i].Date == records[j].Date {
				return records[i].Hour < records[j].Hour
			}
			return records[i].Date < records[j].Date
		})
		for _, record := range records {
			recordSlice := []string{record.Date, record.Hour, fmt.Sprintf("%.2f", record.KWhValue)}
			if err := writer.Write(recordSlice); err != nil {
				return err
			}
		}
	case map[string]float64:
		// Convert data to a slice for sorting
		var records []DailyData
		for date, value := range v {
			records = append(records, DailyData{
				Date:     date,
				KWhValue: value,
			})
		}
		// Sort records by date
		sort.Slice(records, func(i, j int) bool {
			return records[i].Date < records[j].Date
		})
		for _, record := range records {
			recordSlice := []string{record.Date, fmt.Sprintf("%.2f", record.KWhValue)}
			if err := writer.Write(recordSlice); err != nil {
				return err
			}
		}
	}
	return nil
}

func saveHourlyCSV(data map[string]map[string]float64) {
	filePath := "output_data/hourly_data.csv"
	if err := saveCSV(filePath, data, []string{"Date", "Hour", "kWh Value"}); err != nil {
		fmt.Println("Error saving hourly CSV:", err)
	}
}

func saveDailyCSV(data map[string]float64) {
	filePath := "output_data/daily_data.csv"
	if err := saveCSV(filePath, data, []string{"Date", "kWh Value"}); err != nil {
		fmt.Println("Error saving daily CSV:", err)
	}
}

// printData function to display the aggregated data in tabular format
func printData(hourlyData map[string]map[string]float64, dailyData map[string]float64) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	fmt.Fprintln(w, "Hourly Data:")
	fmt.Fprintln(w, "Date\tHour\tkWh Value")
	// Prepare hourly data for printing
	var hourlyRecords []HourlyData
	for date, hours := range hourlyData {
		for hour, value := range hours {
			hourlyRecords = append(hourlyRecords, HourlyData{
				Date:     date,
				Hour:     hour,
				KWhValue: value,
			})
		}
	}
	// Sort hourly records by date and hour
	sort.Slice(hourlyRecords, func(i, j int) bool {
		if hourlyRecords[i].Date == hourlyRecords[j].Date {
			return hourlyRecords[i].Hour < hourlyRecords[j].Hour
		}
		return hourlyRecords[i].Date < hourlyRecords[j].Date
	})
	for _, record := range hourlyRecords {
		fmt.Fprintf(w, "%s\t%s\t%.2f\n", record.Date, record.Hour, record.KWhValue)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Daily Data:")
	fmt.Fprintln(w, "Date\tkWh Value")
	// Prepare daily data for printing
	var dailyRecords []DailyData
	for date, value := range dailyData {
		dailyRecords = append(dailyRecords, DailyData{
			Date:     date,
			KWhValue: value,
		})
	}
	// Sort daily records by date
	sort.Slice(dailyRecords, func(i, j int) bool {
		return dailyRecords[i].Date < dailyRecords[j].Date
	})
	for _, record := range dailyRecords {
		fmt.Fprintf(w, "%s\t%.2f\n", record.Date, record.KWhValue)
	}
	w.Flush()
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

	// Uncomment to Print data to the console
	// printData(hourlyData, dailyData)

	// Save data sequentially
	saveHourlyCSV(hourlyData)
	saveDailyCSV(dailyData)

	fmt.Println("Data saved successfully.")
}
