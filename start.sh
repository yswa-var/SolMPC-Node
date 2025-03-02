#!/bin/sh
nats-server -DV &
sleep 2
./tilt-validator start "$@"
