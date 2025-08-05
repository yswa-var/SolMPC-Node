package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TestResults represents the results of a test run
type TestResults struct {
	TestName      string    `json:"test_name"`
	Status        string    `json:"status"`
	Duration      string    `json:"duration"`
	Output        string    `json:"output"`
	Error         string    `json:"error,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	CoverageData  string    `json:"coverage_data,omitempty"`
	BenchmarkData string    `json:"benchmark_data,omitempty"`
}

// TestSuite represents a collection of test results
type TestSuite struct {
	SuiteName    string        `json:"suite_name"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	SkippedTests int           `json:"skipped_tests"`
	Results      []TestResults `json:"results"`
}

// TestRunner manages the execution of all MPC test suites
type TestRunner struct {
	verbose        bool
	generateReport bool
	runBenchmarks  bool
	outputDir      string
	packagePath    string
}

func main() {
	runner := &TestRunner{}

	// Parse command line flags
	flag.BoolVar(&runner.verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&runner.generateReport, "report", true, "Generate detailed test report")
	flag.BoolVar(&runner.runBenchmarks, "bench", false, "Run benchmark tests")
	flag.StringVar(&runner.outputDir, "output", "./test-reports", "Output directory for test reports")
	flag.StringVar(&runner.packagePath, "pkg", "./internal/mpc", "Package path for tests")
	flag.Parse()

	fmt.Println("🚀 Starting End-to-End MPC Flow Verification Test Suite")
	fmt.Println(strings.Repeat("=", 80))

	if err := runner.run(); err != nil {
		fmt.Printf("❌ Test suite failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Test suite completed successfully!")
}

func (tr *TestRunner) run() error {
	// Create output directory
	if err := os.MkdirAll(tr.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	testSuite := &TestSuite{
		SuiteName: "MPC End-to-End Verification",
		StartTime: time.Now(),
	}

	// Run integration tests
	fmt.Println("\n📋 Running Integration Tests...")
	integrationResults, err := tr.runIntegrationTests()
	if err != nil {
		return fmt.Errorf("integration tests failed: %w", err)
	}
	testSuite.Results = append(testSuite.Results, integrationResults...)

	// Run benchmark tests if requested
	if tr.runBenchmarks {
		fmt.Println("\n⚡ Running Benchmark Tests...")
		benchmarkResults, err := tr.runBenchmarkTests()
		if err != nil {
			fmt.Printf("⚠️ Benchmark tests failed: %v\n", err)
		} else {
			testSuite.Results = append(testSuite.Results, benchmarkResults...)
		}
	}

	// Run coverage analysis
	fmt.Println("\n📊 Running Coverage Analysis...")
	coverageResults, err := tr.runCoverageTests()
	if err != nil {
		fmt.Printf("⚠️ Coverage analysis failed: %v\n", err)
	} else {
		testSuite.Results = append(testSuite.Results, coverageResults...)
	}

	// Finalize test suite
	testSuite.EndTime = time.Now()
	testSuite.Duration = testSuite.EndTime.Sub(testSuite.StartTime)

	// Calculate statistics
	for _, result := range testSuite.Results {
		testSuite.TotalTests++
		switch result.Status {
		case "PASS":
			testSuite.PassedTests++
		case "FAIL":
			testSuite.FailedTests++
		case "SKIP":
			testSuite.SkippedTests++
		}
	}

	// Generate reports
	if tr.generateReport {
		if err := tr.generateTestReport(testSuite); err != nil {
			return fmt.Errorf("failed to generate test report: %w", err)
		}
	}

	tr.printSummary(testSuite)

	if testSuite.FailedTests > 0 {
		return fmt.Errorf("test suite failed with %d failed tests", testSuite.FailedTests)
	}

	return nil
}

func (tr *TestRunner) runIntegrationTests() ([]TestResults, error) {
	var results []TestResults

	tests := []string{
		"TestEndToEndMPCFlow",
		"TestErrorHandling",
		"TestPerformanceMetrics",
		"TestSecurityAudit",
	}

	for _, test := range tests {
		fmt.Printf("  🧪 Running %s...\n", test)

		result := tr.runSingleTest(test, "integration")
		results = append(results, result)

		if result.Status == "FAIL" {
			fmt.Printf("    ❌ %s failed\n", test)
		} else {
			fmt.Printf("    ✅ %s passed (%s)\n", test, result.Duration)
		}
	}

	return results, nil
}

func (tr *TestRunner) runBenchmarkTests() ([]TestResults, error) {
	var results []TestResults

	benchmarks := []string{
		"BenchmarkDKG",
		"BenchmarkMPCSigning",
		"BenchmarkFullFlow",
		"BenchmarkValidatorScaling",
		"BenchmarkMessageThroughput",
		"BenchmarkMemoryUsage",
	}

	for _, bench := range benchmarks {
		fmt.Printf("  ⚡ Running %s...\n", bench)

		result := tr.runBenchmark(bench)
		results = append(results, result)

		if result.Status == "FAIL" {
			fmt.Printf("    ❌ %s failed\n", bench)
		} else {
			fmt.Printf("    ✅ %s completed (%s)\n", bench, result.Duration)
		}
	}

	return results, nil
}

func (tr *TestRunner) runCoverageTests() ([]TestResults, error) {
	fmt.Printf("  📊 Running test coverage analysis...\n")

	result := tr.runSingleTest("", "coverage")

	if result.Status == "FAIL" {
		fmt.Printf("    ❌ Coverage analysis failed\n")
	} else {
		fmt.Printf("    ✅ Coverage analysis completed\n")
	}

	return []TestResults{result}, nil
}

func (tr *TestRunner) runSingleTest(testName, testType string) TestResults {
	startTime := time.Now()

	var cmd *exec.Cmd
	var result TestResults

	switch testType {
	case "integration":
		args := []string{"test"}
		if tr.verbose {
			args = append(args, "-v")
		}
		args = append(args, "-run", testName, tr.packagePath)
		cmd = exec.Command("go", args...)

	case "coverage":
		args := []string{"test", "-coverprofile=coverage.out", "-covermode=atomic"}
		if tr.verbose {
			args = append(args, "-v")
		}
		args = append(args, tr.packagePath)
		cmd = exec.Command("go", args...)
		testName = "Coverage Analysis"

	default:
		result.Status = "SKIP"
		result.Error = "Unknown test type"
		return result
	}

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	result = TestResults{
		TestName:  testName,
		Duration:  duration.String(),
		Output:    string(output),
		Timestamp: startTime,
	}

	if err != nil {
		result.Status = "FAIL"
		result.Error = err.Error()
	} else {
		result.Status = "PASS"
	}

	// Handle coverage output
	if testType == "coverage" && result.Status == "PASS" {
		coverageOutput, _ := exec.Command("go", "tool", "cover", "-func=coverage.out").CombinedOutput()
		result.CoverageData = string(coverageOutput)
	}

	return result
}

func (tr *TestRunner) runBenchmark(benchName string) TestResults {
	startTime := time.Now()

	args := []string{"test", "-bench", benchName, "-benchmem"}
	if tr.verbose {
		args = append(args, "-v")
	}
	args = append(args, tr.packagePath)

	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	result := TestResults{
		TestName:      benchName,
		Duration:      duration.String(),
		Output:        string(output),
		BenchmarkData: string(output),
		Timestamp:     startTime,
	}

	if err != nil {
		result.Status = "FAIL"
		result.Error = err.Error()
	} else {
		result.Status = "PASS"
	}

	return result
}

func (tr *TestRunner) generateTestReport(testSuite *TestSuite) error {
	// Generate JSON report
	jsonData, err := json.MarshalIndent(testSuite, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test results: %w", err)
	}

	jsonFile := filepath.Join(tr.outputDir, "test-results.json")
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	// Generate HTML report
	htmlReport := tr.generateHTMLReport(testSuite)
	htmlFile := filepath.Join(tr.outputDir, "test-report.html")
	if err := os.WriteFile(htmlFile, []byte(htmlReport), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	// Generate markdown summary
	markdownReport := tr.generateMarkdownReport(testSuite)
	mdFile := filepath.Join(tr.outputDir, "test-summary.md")
	if err := os.WriteFile(mdFile, []byte(markdownReport), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report: %w", err)
	}

	fmt.Printf("📄 Test reports generated in: %s\n", tr.outputDir)
	return nil
}

func (tr *TestRunner) generateHTMLReport(testSuite *TestSuite) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>MPC Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .test-result { margin: 10px 0; padding: 10px; border-radius: 5px; }
        .pass { background: #e8f5e8; border-left: 5px solid #4caf50; }
        .fail { background: #ffeaea; border-left: 5px solid #f44336; }
        .skip { background: #fff3cd; border-left: 5px solid #ffc107; }
        .details { margin-top: 10px; font-family: monospace; background: #f8f8f8; padding: 10px; border-radius: 3px; }
        pre { white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧪 MPC End-to-End Verification Test Results</h1>
        <p><strong>Suite:</strong> ` + testSuite.SuiteName + `</p>
        <p><strong>Start Time:</strong> ` + testSuite.StartTime.Format(time.RFC3339) + `</p>
        <p><strong>Duration:</strong> ` + testSuite.Duration.String() + `</p>
    </div>
    
    <div class="summary">
        <h2>📊 Summary</h2>
        <p>Total Tests: ` + strconv.Itoa(testSuite.TotalTests) + `</p>
        <p>✅ Passed: ` + strconv.Itoa(testSuite.PassedTests) + `</p>
        <p>❌ Failed: ` + strconv.Itoa(testSuite.FailedTests) + `</p>
        <p>⏭️ Skipped: ` + strconv.Itoa(testSuite.SkippedTests) + `</p>
    </div>
    
    <h2>🧪 Test Results</h2>`

	for _, result := range testSuite.Results {
		cssClass := strings.ToLower(result.Status)
		html += `
    <div class="test-result ` + cssClass + `">
        <h3>` + result.TestName + ` - ` + result.Status + `</h3>
        <p><strong>Duration:</strong> ` + result.Duration + `</p>`

		if result.Error != "" {
			html += `<p><strong>Error:</strong> ` + result.Error + `</p>`
		}

		if result.Output != "" {
			html += `<div class="details"><strong>Output:</strong><pre>` + result.Output + `</pre></div>`
		}

		html += `</div>`
	}

	html += `</body></html>`
	return html
}

func (tr *TestRunner) generateMarkdownReport(testSuite *TestSuite) string {
	md := fmt.Sprintf(`# 🧪 MPC End-to-End Verification Test Results

## 📋 Test Suite Information
- **Suite Name:** %s
- **Start Time:** %s
- **Duration:** %s

## 📊 Summary
- **Total Tests:** %d
- **✅ Passed:** %d
- **❌ Failed:** %d
- **⏭️ Skipped:** %d
- **Success Rate:** %.1f%%

## 🧪 Detailed Results

`, testSuite.SuiteName,
		testSuite.StartTime.Format(time.RFC3339),
		testSuite.Duration.String(),
		testSuite.TotalTests,
		testSuite.PassedTests,
		testSuite.FailedTests,
		testSuite.SkippedTests,
		float64(testSuite.PassedTests)/float64(testSuite.TotalTests)*100)

	for _, result := range testSuite.Results {
		status := "✅"
		if result.Status == "FAIL" {
			status = "❌"
		} else if result.Status == "SKIP" {
			status = "⏭️"
		}

		md += fmt.Sprintf("### %s %s\n", status, result.TestName)
		md += fmt.Sprintf("- **Status:** %s\n", result.Status)
		md += fmt.Sprintf("- **Duration:** %s\n", result.Duration)

		if result.Error != "" {
			md += fmt.Sprintf("- **Error:** %s\n", result.Error)
		}

		md += "\n"
	}

	return md
}

func (tr *TestRunner) printSummary(testSuite *TestSuite) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 TEST SUITE SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("Suite: %s\n", testSuite.SuiteName)
	fmt.Printf("Duration: %s\n", testSuite.Duration)
	fmt.Printf("Total Tests: %d\n", testSuite.TotalTests)
	fmt.Printf("✅ Passed: %d\n", testSuite.PassedTests)
	fmt.Printf("❌ Failed: %d\n", testSuite.FailedTests)
	fmt.Printf("⏭️ Skipped: %d\n", testSuite.SkippedTests)

	if testSuite.TotalTests > 0 {
		successRate := float64(testSuite.PassedTests) / float64(testSuite.TotalTests) * 100
		fmt.Printf("Success Rate: %.1f%%\n", successRate)
	}

	fmt.Println(strings.Repeat("=", 80))

	if testSuite.FailedTests == 0 {
		fmt.Println("🎉 ALL TESTS PASSED!")
	} else {
		fmt.Printf("⚠️ %d TESTS FAILED\n", testSuite.FailedTests)
		fmt.Println("\nFailed tests:")
		for _, result := range testSuite.Results {
			if result.Status == "FAIL" {
				fmt.Printf("  ❌ %s: %s\n", result.TestName, result.Error)
			}
		}
	}
}
