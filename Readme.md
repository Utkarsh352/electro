# Electro - Analytics Service

## 1. Introduction

Electro is an analytics service designed to process and aggregate kilowatt-hour (kWh) data. This service receives a dataset with kWh values and timestamps, and transforms it into meaningful hourly and daily summaries for further analysis or reporting.

## 2. Objective

The primary objective of the Electro service is to:
1. Aggregate kWh values into hourly segments.
2. Aggregate kWh values into daily segments.

## 3. Scope

The service includes:
- Processing kWh data with timestamps.
- Aggregating data from multiple sources or meters.
- Generating hourly and daily aggregated kWh values.

## 4. Requirements

### 4.1 Input Data

- **Dataset Structure:**
  - Each record should include:
    - `timestamp`: The date and time when the kWh value was recorded.
    - `kWh_value`: The kilowatt-hour value recorded at the given timestamp.
  - Optionally, a `source_id` or `meter_id` may be included to differentiate data sources.

- **Timestamp Format:**
  - Timestamps must be in ISO 8601 format (e.g., `YYYY-MM-DDTHH:MM:SSZ`).

### 4.2 Processing Logic

- **Time Zone Handling:**
  - The service operates in UTC time zone by default.

- **Hourly Aggregation:**
  - kWh values are grouped by hour, summing all values within the same hour.
  - Output includes a timestamp representing the start of each hour and the aggregated kWh value.

- **Daily Aggregation:**
  - kWh values are grouped by day, summing all values within the same day.
  - Output includes a date representing the day and the aggregated kWh value.

- **Missing Data Handling:**
  - If no data exists for a particular hour or day, the service returns a record with a `0` kWh value for that period.

- **Overlapping Data:**
  - Overlapping or duplicate timestamps are aggregated into the appropriate hourly or daily segment.

### 4.3 Output Data

- **Hourly Aggregated Output:**
  - Dataset containing:
    - `hourly_timestamp`: Start time of each hour.
    - `aggregated_kWh_value`: Sum of kWh values within that hour.

- **Daily Aggregated Output:**
  - Dataset containing:
    - `daily_timestamp`: Date for each day.
    - `aggregated_kWh_value`: Sum of kWh values within that day.

- **Output Format:**
  - Output is generated in CSV format.

## 5. Implementation

### 5.1 Code Overview

- **Data Parsing:**
  - Parses the input JSON data to extract timestamps and kWh values.

- **Data Aggregation:**
  - Aggregates kWh values into hourly and daily summaries based on the parsed timestamps.

- **CSV File Creation:**
  - Saves hourly and daily aggregated data into separate CSV files in the `output_data` directory.

### 5.2 File Operations

- **Hourly Data:**
  - Each date’s hourly data is saved into a CSV file named `hourly_data_<date>.csv`, where `<date>` is formatted as `DD_MM`.

- **Daily Data:**
  - Daily aggregated data is saved into a single CSV file named `daily_data.csv`.

### 5.3 Sequential Execution

- The service performs all file operations sequentially, without concurrency, to simplify file handling and avoid potential race conditions.

## 6. Non-Functional Requirements

- **Performance:**
  - The service processes data efficiently and completes aggregation within a reasonable timeframe.

- **Scalability:**
  - Designed to handle increasing data volumes from multiple sources.

- **Reliability:**
  - Ensures data integrity during aggregation.

- **Error Handling:**
  - Logs errors encountered during processing and provides meaningful error messages.


** to print output, uncomment the print line in the code **