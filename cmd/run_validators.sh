#!/bin/bash

# If inside tmux, exit to prevent nesting issues
if [ -n "$TMUX" ]; then
    echo "You're already inside a tmux session. Run this script outside tmux."
    exit 1
fi

# Create a new tmux session named "validators" and start the first instance
tmux new-session -d -s validators "go run . 1; exec bash"

# Split the window horizontally and start the second instance
tmux split-window -h "go run . 2; exec bash"

# Split the right pane vertically and start the third instance
tmux split-window -v "go run . 3; exec bash"

# Select the first pane for better focus
tmux select-pane -t 0

# Attach to the session so the user sees the instances running
tmux attach -t validators
