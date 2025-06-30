#!/bin/bash
echo -e "\\bash echo 'Test 1 works!'\n\\bash pwd\n\\bash ls | head -2\n\\exit" | ./bin/neuro shell