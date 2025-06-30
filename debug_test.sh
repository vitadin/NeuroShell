#!/bin/bash
echo -e "\\bash echo 'Test 1'\n\\bash ls | head -2\n\\exit" | ./bin/neuro --log-level=debug shell 2>&1 | head -50