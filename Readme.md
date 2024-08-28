Here is the documentation for the code named "Electro":

---

# Electro: KWh Data Aggregation Service

## Introduction

Electro is an analytics service designed to process and aggregate kilowatt-hour (kWh) data. This service takes raw kWh data with timestamps, and generates hourly and daily summaries. It outputs the aggregated data into CSV files for further analysis or reporting.

## Objective

The primary objective of Electro is to:

1. Aggregate kWh values into hourly segments.
2. Aggregate kWh values into daily segments.

## Scope

Electro handles the following tasks:

- Processing kWh data with timestamps.
- Aggregating data from multiple sources or meters.
- Generating both hourly and daily aggregated kWh values.

## Requirements

### Input Data

- **Dataset Structure:**
  - Each record must include:
    - `timestamp`: The date and time when the kWh value was recorded (ISO 8601 format).
    - `kWh_value`: The kilowatt-hour value recorded at the given timestamp.
  - Optionally, records may include `source_id` or `meter_id`.

- **Timestamp Format:**
  - Must be in ISO 8601 format (e.g., `YYYY-MM-DDTHH:MM:SSZ`).

### Processing Logic

- **Time Zone Handling:**
  - Handles data in multiple time zones. The default time zone is set to UTC.

- **Hourly Aggregation:**
  - Groups kWh values by hour, summing all values within the same hour (e.g., 01:00:00 to 01:59:59).
  - Outputs include a timestamp representing the start of the hour and the aggregated kWh value.

- **Daily Aggregation:**
  - Groups kWh values by day, summing all values within the same day (e.g., 00:00:00 to 23:59:59).
  - Outputs include a date representing the day and the aggregated kWh value.

- **Missing Data Handling:**
  - If no data is present for a particular hour or day, the output will show a `0` kWh value for that period.

- **Overlapping Data:**
  - Aggregates overlapping or duplicate timestamps into the appropriate hourly or daily segment.

### Output Data

- **Hourly Aggregated Output:**
  - CSV files for each day, containing:
    - `Date`: The date for the data.
    - `Hour`: The start time of each hour.
    - `kWh Value`: The sum of kWh values within that hour.

- **Daily Aggregated Output:**
  - A single CSV file containing:
    - `Date`: The date for each day in `dd/mm` format.
    - `kWh Value`: The sum of kWh values within that day.

### File Output Directory

- **Output Directory:**
  - The output CSV files are stored in a directory named `output_data`.

## Installation

1. Clone the repository or download the source code.
2. Navigate to the project directory.
3. Ensure you have Go installed on your system.
4. Run the following command to execute the code:

   ```sh
   go run main.go
   ```

## Code Structure

### Main Components

- **Data Structures:**
  - `Record`: Defines the structure of the input data.
  - `HourlyData` and `DailyData`: Define the structure of the output data for CSV files.

- **Functions:**
  - `parseTimestamp(ts string, loc *time.Location) (time.Time, error)`: Parses the timestamp in the specified time zone.
  - `aggregateData(records []Record, loc *time.Location) (map[string]map[string]float64, map[string]float64)`: Aggregates the data into hourly and daily segments.
  - `saveCSV(filePath string, data interface{}, headers []string) error`: Saves data to a CSV file.
  - `saveHourlyCSV(data map[string]map[string]float64, wg *sync.WaitGroup)`: Saves hourly aggregated data to CSV files.
  - `saveDailyCSV(data map[string]float64, wg *sync.WaitGroup)`: Saves daily aggregated data to a CSV file.
  - `main()`: The entry point of the application. Reads input data, processes it, and saves the output.

## Error Handling

- The application logs any errors encountered during processing.
- It ensures that data integrity is maintained throughout the aggregation process.

## License

This project is licensed under the MIT License.

## Contact

For any questions or issues, please contact the maintainer at [email@example.com](mailto:email@example.com).

---

This documentation provides an overview of the Electro service, its requirements, installation instructions, and how it processes and outputs kWh data.