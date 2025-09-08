#!/usr/bin/env bash

# Test the specific failing case
echo "=== Testing specific failing case ==="

# Use printf to avoid shell quoting issues
printf '\send Hello! Can you help me understand these symbols: @#$%%^&*()_+-=[]{}|;'\\":\\",./<>?~ and émojis like café, résumé, naïve?\n' | ./bin/neuro 2>&1 | head -10

echo ""
echo "=== Test completed ==="