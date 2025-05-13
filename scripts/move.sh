#!/bin/bash

step=50

curl -X POST -d "reset
green
figure 400 200
update" http://localhost:17000/

while true; do
    curl -X POST http://localhost:17000 -d "update"
    for ((i = 0; i < 500; i += step)); do
        curl -X POST http://localhost:17000 -d "move 0 $((step))"
        curl -X POST http://localhost:17000 -d "update"
    done
    for ((i = 500; i > 0; i -= step)); do
        curl -X POST http://localhost:17000 -d "move 0 $((-step))"
        curl -X POST http://localhost:17000 -d "update"
    done
done