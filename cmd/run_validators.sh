#!/bin/bash

# If inside tmux, exit to prevent nesting issues
if [ -n "$TMUX" ]; then
    echo "You're already inside a tmux session. Run this script outside tmux."
    exit 1
fi

# --tilt-type=simple is used to make the validators tilt easily
# Create a new tmux session named "validators" and start the first instance
# tmux new-session -d -s validators "go run . 1 --tilt-type=simple; exec bash"

# # Split the window horizontally and start the second instance
# tmux split-window -h "go run . 2 --tilt-type=simple; exec bash"

# # Split the right pane vertically and start the third instance
# tmux split-window -v "go run . 3 --tilt-type=simple; exec bash"

# --tilt-type=two_subtilts is used to make the validators tilt easily
# Create a new tmux session named "validators" and start the first instance
tmux new-session -d -s validators "go run . 1 --tilt-type=two_subtilts; exec bash"

# Split the window horizontally and start the second instance
tmux split-window -h "go run . 2 --tilt-type=two_subtilts; exec bash"

# Split the right pane vertically and start the third instance
tmux split-window -v "go run . 3 --tilt-type=two_subtilts; exec bash"

# Select the first pane for better focus
tmux select-pane -t 0

# Attach to the session so the user sees the instances running
tmux attach -t validators
