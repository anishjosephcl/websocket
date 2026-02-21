package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
)

// --- 1. Data Structures ---

// Transformation defines a single transformation rule.
type Transformation struct {
	ColumnName string
	FuncName   string // e.g., "Change"
	Parameter  string
}

// ProcessedChunk is used to send results from goroutines over the channel.
// It includes the original index to reassemble chunks in the correct order.
type ProcessedChunk struct {
	Index int
	Data  [][]string
}

// CSVProcessor encapsulates the logic for processing a CSV file.
type CSVProcessor struct {
	InputFile  string
	OutputFile string
	Transforms []Transformation
}

// NewCSVProcessor is a constructor for our main struct.
func NewCSVProcessor(inputFile, outputFile string) *CSVProcessor {
	return &CSVProcessor{
		InputFile:  inputFile,
		OutputFile: outputFile,
		Transforms: []Transformation{},
	}
}

// AddTransform is a helper to add new transformation rules.
func (p *CSVProcessor) AddTransform(column, funcName, param string) {
	p.Transforms = append(p.Transforms, Transformation{
		ColumnName: column,
		FuncName:   funcName,
		Parameter:  param,
	})
}

// --- 2. Core Functionality ---

// GetMetadata reads the CSV and returns the headers and the number of data rows.
func (p *CSVProcessor) GetMetadata() ([]string, int, error) {
	file, err := os.Open(p.InputFile)
	if err != nil {
		return nil, 0, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("could not read CSV data: %w", err)
	}

	if len(records) == 0 {
		return nil, 0, fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	numDataLines := len(records) - 1

	return headers, numDataLines, nil
}

// chunkData breaks the CSV records into smaller chunks of a given size.
func chunkData(records [][]string, chunkSize int) [][][]string {
	var chunks [][][]string
	for i := 0; i < len(records); i += chunkSize {
		end := i + chunkSize
		if end > len(records) {
			end = len(records)
		}
		chunks = append(chunks, records[i:end])
	}
	return chunks
}

// applyTransformations is the core logic executed by each goroutine on its chunk.
func applyTransformations(chunk [][]string, headerMap map[string]int, transforms []Transformation) [][]string {
	// Create a map of which column index needs which transformation.
	// This is more efficient than searching the header for every row.
	colIndexToTransform := make(map[int]Transformation)
	for _, t := range transforms {
		if colIndex, ok := headerMap[t.ColumnName]; ok {
			colIndexToTransform[colIndex] = t
		}
	}

	// Create a new slice for the transformed data to avoid modifying the original chunk data.
	transformedChunk := make([][]string, len(chunk))
	for i, row := range chunk {
		newRow := make([]string, len(row))
		copy(newRow, row)
		for colIndex, transform := range colIndexToTransform {
			if colIndex < len(newRow) {
				switch transform.FuncName {
				case "Change":
					newRow[colIndex] = transform.Parameter + newRow[colIndex]
				// Future transform functions could be added here.
				default:
					// Do nothing if function name is not recognized.
				}
			}
		}
		transformedChunk[i] = newRow
	}
	return transformedChunk
}

// Process orchestrates the entire workflow.
func (p *CSVProcessor) Process() error {
	// 1. Read all data from the input file.
	file, err := os.Open(p.InputFile)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	allRecords, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read CSV data: %w", err)
	}
	if len(allRecords) < 1 {
		return fmt.Errorf("CSV has no data")
	}

	headers := allRecords[0]
	dataRecords := allRecords[1:]

	// 2. Create a header map for efficient column lookups.
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	// 3. Chunk the data for parallel processing.
	chunks := chunkData(dataRecords, 10)
	log.Printf("Split data into %d chunks of size 10.\n", len(chunks))

	// 4. Set up concurrency tools.
	var wg sync.WaitGroup
	// Unbuffered channel, as requested. Goroutines will block until the main thread receives the result.
	resultsChan := make(chan ProcessedChunk)

	// 5. Spawn a goroutine for each chunk.
	for i, chunk := range chunks {
		wg.Add(1)
		go func(chunkIndex int, chunkData [][]string) {
			defer wg.Done()
			log.Printf("Goroutine %d: Starting processing.\n", chunkIndex)

			transformedData := applyTransformations(chunkData, headerMap, p.Transforms)

			log.Printf("Goroutine %d: Finished processing. Sending to channel.\n", chunkIndex)
			resultsChan <- ProcessedChunk{Index: chunkIndex, Data: transformedData}
		}(i, chunk)
	}

	// 6. Start a goroutine to wait for all workers to finish and then close the channel.
	// This is a crucial pattern to safely range over the channel.
	go func() {
		wg.Wait()
		close(resultsChan)
		log.Println("All goroutines finished. Channel closed.")
	}()

	// 7. Collect results from the channel and reassemble them in order.
	log.Println("Main thread: Waiting for results from channel...")
	// We create a slice to hold the processed chunks in their original order.
	orderedResults := make([][][]string, len(chunks))
	for result := range resultsChan {
		log.Printf("Main thread: Received chunk %d from channel.\n", result.Index)
		orderedResults[result.Index] = result.Data
	}

	// 8. Write the final result to the output file.
	outputFile, err := os.Create(p.OutputFile)
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write header first.
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write all the processed and re-ordered records.
	for _, chunk := range orderedResults {
		if err := writer.WriteAll(chunk); err != nil {
			return fmt.Errorf("failed to write data chunk: %w", err)
		}
	}

	log.Printf("Successfully wrote transformed data to %s\n", p.OutputFile)
	return nil
}

// --- 3. Main Function to Drive the Program ---

func main() {
	// --- Setup: Create a dummy CSV file for the demonstration ---
	inputFile := "input.csv"
	createDummyCSV(inputFile)

	// --- Usage ---
	// 1. Create a processor.
	processor := NewCSVProcessor(inputFile, "output.csv")

	// 2. Get and print metadata.
	headers, lineCount, err := processor.GetMetadata()
	if err != nil {
		log.Fatalf("Failed to get metadata: %v", err)
	}
	log.Printf("--- File Metadata ---\nHeaders: %v\nData Lines: %d\n--------------------\n", headers, lineCount)

	// 3. Define the transformations.
	processor.AddTransform("Department", "Change", "DEPT-")
	processor.AddTransform("EmployeeID", "Change", "EMP_")

	// 4. Run the main processing workflow.
	if err := processor.Process(); err != nil {
		log.Fatalf("Processing failed: %v", err)
	}
}

// createDummyCSV is a helper to generate a sample input file.
func createDummyCSV(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create dummy CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"EmployeeID", "Name", "Department", "Location"}
	writer.Write(headers)

	// Create 25 records to demonstrate chunking.
	for i := 1; i <= 25; i++ {
		record := []string{
			fmt.Sprintf("E%d", 100+i),
			fmt.Sprintf("User %d", i),
			"Engineering",
			"New York",
		}
		if i > 15 {
			record[2] = "Sales"
			record[3] = "London"
		}
		writer.Write(record)
	}
}
