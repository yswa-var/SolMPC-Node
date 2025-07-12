# Contributing to Tilt-Valid

Thank you for your interest in contributing! Here’s a quick guide to help you get started.

## Setup

1. Clone the repository:
   ```bash
   git clone <repo-url>
   cd tilt-validator
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Run the system locally using the provided scripts (e.g., `cmd/run_validators.sh`).

## Key Files & Directories

- `cmd/main.go`: Main application logic and entry point
- `internal/mpc/`: Multi-party computation (MPC) logic
- `internal/exchange/`: File-based transport layer
- `internal/distribution/`: Payment distribution logic
- `utils/`: Utility functions and tilt data helpers
- `data/validators.csv`: Validator configuration

## How to Contribute

1. Fork the repository and create a new branch for your feature or bugfix.
2. Make your changes with clear, concise commits.
3. Add or update tests if applicable.
4. Ensure your code follows Go best practices and is well-documented.
5. Open a pull request with a description of your changes.

## Coding Style

- Use `gofmt` to format your code.
- Write clear, descriptive commit messages.
- Keep functions small and focused.
- Add comments where necessary for clarity.

## Need Help?

- Open an issue in the repository for questions or suggestions.
- For urgent help, contact the maintainers listed in the repository.

Happy coding!
