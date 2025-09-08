#!/usr/bin/env bash

# Test script to isolate the issue with special characters
echo "Testing special characters in send command..."

# First, let's test if the issue is with variable interpolation
echo "=== Testing variable interpolation ==="
go run -c 'println("Testing interpolation:", "${@#$%^&*()_+-=[]{}|;:\",./<>?~}")'

# Test direct send command with special characters
echo "=== Testing direct send command ==="
echo '\send Hello! Can you help me understand these symbols: @#$%^&*()_+-=[]{}|;:'\",./<>?~ and émojis like café, résumé, naïve?' | ./bin/neuro

echo "Done."