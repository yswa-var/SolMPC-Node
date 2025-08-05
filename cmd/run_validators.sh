#!/bin/bash

# SolMPC-Node Validators Runner
# Runs 3 validator instances for MPC threshold signing and ballot processing

# If inside tmux, exit to prevent nesting issues
if [ -n "$TMUX" ]; then
    echo "You're already inside a tmux session. Run this script outside tmux."
    exit 1
fi

# Check if a validators session already exists and kill it
if tmux has-session -t validators 2>/dev/null; then
    echo "Killing existing validators session..."
    tmux kill-session -t validators
fi

echo "Starting 3 MPC validators for ballot processing..."
echo "Each validator participates in:"
echo "  - Distributed Key Generation (DKG)"
echo "  - MPC threshold transaction signing"  
echo "  - Voting ballot processing"
echo "  - VRF-based validator selection"
echo ""

# Create a new tmux session named "validators" and start the first instance
tmux new-session -d -s validators "echo 'Starting Validator 1...'; go run *.go 1; exec bash"

# Add a small delay to prevent race conditions
sleep 1

# Split the window horizontally and start the second instance
tmux split-window -h "echo 'Starting Validator 2...'; go run *.go 2; exec bash"

# Add a small delay
sleep 1

# Split the right pane vertically and start the third instance
tmux split-window -v "echo 'Starting Validator 3...'; go run *.go 3; exec bash"

# Select the first pane for better focus
tmux select-pane -t 0

# Set up layout for better viewing
tmux select-layout even-horizontal

echo "Validators started in tmux session 'validators'"
echo "Use 'tmux kill-session -t validators' to stop all validators"
echo ""

# Attach to the session so the user sees the instances running
tmux attach -t validators
