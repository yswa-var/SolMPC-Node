#!/bin/bash

# MPC End-to-End Test Runner Script
# This script runs the comprehensive test suite for the MPC voting system

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_OUTPUT_DIR="${PROJECT_ROOT}/test-reports"
COVERAGE_FILE="${TEST_OUTPUT_DIR}/coverage.out"
LOG_FILE="${TEST_OUTPUT_DIR}/test-run.log"

# Default options
VERBOSE=false
RUN_BENCHMARKS=false
GENERATE_REPORTS=true
CLEAN_REPORTS=false
RUN_INTEGRATION=true
RUN_SECURITY=true
RUN_PERFORMANCE=true

# Parse command line arguments
usage() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  -v, --verbose          Enable verbose output"
    echo "  -b, --benchmarks       Run benchmark tests"
    echo "  -r, --no-reports       Skip report generation"
    echo "  -c, --clean            Clean previous test reports"
    echo "  --integration-only     Run only integration tests"
    echo "  --security-only        Run only security tests"
    echo "  --performance-only     Run only performance tests"
    echo "  -h, --help            Show this help message"
    exit 1
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -b|--benchmarks)
            RUN_BENCHMARKS=true
            shift
            ;;
        -r|--no-reports)
            GENERATE_REPORTS=false
            shift
            ;;
        -c|--clean)
            CLEAN_REPORTS=true
            shift
            ;;
        --integration-only)
            RUN_INTEGRATION=true
            RUN_SECURITY=false
            RUN_PERFORMANCE=false
            shift
            ;;
        --security-only)
            RUN_INTEGRATION=false
            RUN_SECURITY=true
            RUN_PERFORMANCE=false
            shift
            ;;
        --performance-only)
            RUN_INTEGRATION=false
            RUN_SECURITY=false
            RUN_PERFORMANCE=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Utility functions
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

separator() {
    echo -e "\n${BLUE}===== $1 =====${NC}\n" | tee -a "$LOG_FILE"
}

# Setup functions
setup_environment() {
    separator "Setting up test environment"
    
    # Change to project root
    cd "$PROJECT_ROOT"
    
    # Create test output directory
    mkdir -p "$TEST_OUTPUT_DIR"
    
    # Clean previous reports if requested
    if [ "$CLEAN_REPORTS" = true ]; then
        log_info "Cleaning previous test reports..."
        rm -rf "${TEST_OUTPUT_DIR:?}"/*
    fi
    
    # Initialize log file
    echo "Test run started at $(date)" > "$LOG_FILE"
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    log_info "Go version: $(go version)"
    
    # Check if we're in a Go module
    if [ ! -f "go.mod" ]; then
        log_error "Not in a Go module directory"
        exit 1
    fi
    
    # Install dependencies
    log_info "Installing/updating dependencies..."
    go mod download
    go mod tidy
    
    log_success "Environment setup complete"
}

# Test execution functions
run_integration_tests() {
    if [ "$RUN_INTEGRATION" != true ]; then
        return 0
    fi
    
    separator "Running Integration Tests"
    
    local test_args="-v"
    if [ "$VERBOSE" = true ]; then
        test_args="$test_args -v"
    fi
    
    log_info "Executing end-to-end MPC flow tests..."
    
    if go test $test_args -run "TestEndToEndMPCFlow" ./internal/mpc 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Integration tests passed"
        return 0
    else
        log_error "Integration tests failed"
        return 1
    fi
}

run_security_audit() {
    if [ "$RUN_SECURITY" != true ]; then
        return 0
    fi
    
    separator "Running Security Audit Tests"
    
    local test_args="-v"
    if [ "$VERBOSE" = true ]; then
        test_args="$test_args -v"
    fi
    
    log_info "Executing comprehensive security audit..."
    
    if go test $test_args -run "TestComprehensiveSecurityAudit" ./internal/mpc 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Security audit tests passed"
        return 0
    else
        log_error "Security audit tests failed"
        return 1
    fi
}

run_performance_tests() {
    if [ "$RUN_PERFORMANCE" != true ]; then
        return 0
    fi
    
    separator "Running Performance Tests"
    
    local test_args="-v"
    if [ "$VERBOSE" = true ]; then
        test_args="$test_args -v"
    fi
    
    log_info "Executing performance and benchmark tests..."
    
    # Run regular performance tests
    if ! go test $test_args -run "Performance" ./internal/mpc 2>&1 | tee -a "$LOG_FILE"; then
        log_error "Performance tests failed"
        return 1
    fi
    
    # Run benchmarks if requested
    if [ "$RUN_BENCHMARKS" = true ]; then
        log_info "Running benchmark tests..."
        if go test -bench=. -benchmem ./internal/mpc 2>&1 | tee -a "$LOG_FILE"; then
            log_success "Benchmark tests completed"
        else
            log_warning "Benchmark tests failed"
        fi
    fi
    
    log_success "Performance tests passed"
    return 0
}

generate_coverage_report() {
    separator "Generating Coverage Report"
    
    log_info "Running tests with coverage analysis..."
    
    # Run all tests with coverage
    if go test -coverprofile="$COVERAGE_FILE" -covermode=atomic ./internal/mpc 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Coverage data generated"
        
        # Generate coverage report
        if command -v go &> /dev/null; then
            log_info "Generating coverage report..."
            go tool cover -html="$COVERAGE_FILE" -o "${TEST_OUTPUT_DIR}/coverage.html"
            go tool cover -func="$COVERAGE_FILE" > "${TEST_OUTPUT_DIR}/coverage.txt"
            
            # Extract coverage percentage
            local coverage_pct=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
            log_info "Total test coverage: $coverage_pct"
            
            # Check coverage threshold (set to 80%)
            local coverage_num=$(echo "$coverage_pct" | sed 's/%//')
            if (( $(echo "$coverage_num >= 80" | bc -l) )); then
                log_success "Coverage threshold met (>= 80%)"
            else
                log_warning "Coverage below threshold: $coverage_pct (expected >= 80%)"
            fi
        fi
    else
        log_error "Coverage analysis failed"
        return 1
    fi
}

run_test_runner() {
    separator "Running Comprehensive Test Runner"
    
    log_info "Building and running custom test runner..."
    
    local runner_args=""
    if [ "$VERBOSE" = true ]; then
        runner_args="$runner_args -v"
    fi
    
    if [ "$RUN_BENCHMARKS" = true ]; then
        runner_args="$runner_args -bench"
    fi
    
    if [ "$GENERATE_REPORTS" = true ]; then
        runner_args="$runner_args -report -output $TEST_OUTPUT_DIR"
    fi
    
    	# Build and run the test runner
	if go run ./cmd/test-runner/main.go $runner_args 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Test runner completed successfully"
        return 0
    else
        log_error "Test runner failed"
        return 1
    fi
}

generate_final_report() {
    if [ "$GENERATE_REPORTS" != true ]; then
        return 0
    fi
    
    separator "Generating Final Test Report"
    
    local report_file="${TEST_OUTPUT_DIR}/final-report.md"
    
    {
        echo "# MPC End-to-End Verification Test Report"
        echo ""
        echo "**Generated:** $(date)"
        echo "**Project:** SolMPC-Node"
        echo ""
        echo "## Test Summary"
        echo ""
        
        # Check if individual test results exist
        if [ -f "${TEST_OUTPUT_DIR}/test-results.json" ]; then
            echo "✅ Comprehensive test suite completed"
            echo ""
            echo "Detailed results available in:"
            echo "- \`test-results.json\` - Machine-readable results"
            echo "- \`test-report.html\` - Human-readable HTML report"
            echo "- \`coverage.html\` - Test coverage visualization"
            echo ""
        fi
        
        echo "## Test Categories"
        echo ""
        if [ "$RUN_INTEGRATION" = true ]; then
            echo "- ✅ **Integration Tests** - End-to-end MPC flow verification"
        fi
        if [ "$RUN_SECURITY" = true ]; then
            echo "- ✅ **Security Audit** - Comprehensive security property verification"
        fi
        if [ "$RUN_PERFORMANCE" = true ]; then
            echo "- ✅ **Performance Tests** - Latency and throughput analysis"
        fi
        if [ "$RUN_BENCHMARKS" = true ]; then
            echo "- ✅ **Benchmark Tests** - Performance benchmarking"
        fi
        
        echo ""
        echo "## Files Generated"
        echo ""
        find "$TEST_OUTPUT_DIR" -type f -name "*.html" -o -name "*.json" -o -name "*.txt" -o -name "*.md" | while read -r file; do
            echo "- \`$(basename "$file")\`"
        done
        
        echo ""
        echo "## Log Output"
        echo ""
        echo "Complete test execution log:"
        echo ""
        echo "\`\`\`"
        tail -50 "$LOG_FILE"
        echo "\`\`\`"
        
    } > "$report_file"
    
    log_success "Final report generated: $report_file"
}

print_summary() {
    separator "Test Execution Summary"
    
    echo -e "${BLUE}📊 Test Execution Results:${NC}"
    echo ""
    
    if [ "$RUN_INTEGRATION" = true ]; then
        echo -e "  ${GREEN}✅${NC} Integration Tests"
    fi
    
    if [ "$RUN_SECURITY" = true ]; then
        echo -e "  ${GREEN}✅${NC} Security Audit Tests"
    fi
    
    if [ "$RUN_PERFORMANCE" = true ]; then
        echo -e "  ${GREEN}✅${NC} Performance Tests"
    fi
    
    if [ "$RUN_BENCHMARKS" = true ]; then
        echo -e "  ${GREEN}✅${NC} Benchmark Tests"
    fi
    
    echo ""
    echo -e "${BLUE}📁 Reports Generated:${NC}"
    if [ -d "$TEST_OUTPUT_DIR" ]; then
        find "$TEST_OUTPUT_DIR" -name "*.html" -o -name "*.json" -o -name "*.md" | while read -r file; do
            echo -e "  📄 $(basename "$file")"
        done
    fi
    
    echo ""
    echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
    echo -e "📂 Results available in: ${CYAN}$TEST_OUTPUT_DIR${NC}"
}

# Main execution
run_main() {
    local start_time=$(date +%s)
    local exit_code=0
    
    echo -e "${CYAN}"
    echo "🚀 MPC End-to-End Verification Test Suite"
    echo "=========================================="
    echo -e "${NC}"
    
    # Setup
    if ! setup_environment; then
        exit 1
    fi
    
    # Run tests
    if ! run_integration_tests; then
        exit_code=1
    fi
    
    if ! run_security_audit; then
        exit_code=1
    fi
    
    if ! run_performance_tests; then
        exit_code=1
    fi
    
    # Generate coverage
    if ! generate_coverage_report; then
        log_warning "Coverage report generation failed"
    fi
    
    # Run comprehensive test runner
    if ! run_test_runner; then
        exit_code=1
    fi
    
    # Generate final report
    generate_final_report
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    separator "Execution Complete"
    log_info "Total execution time: ${duration}s"
    
    if [ $exit_code -eq 0 ]; then
        print_summary
    else
        log_error "Test suite failed with exit code $exit_code"
        echo -e "${RED}❌ Some tests failed. Check the logs for details.${NC}"
    fi
    
    exit $exit_code
}

run_main "$@"