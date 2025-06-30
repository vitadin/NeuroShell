#!/bin/bash
# Interactive test script
echo -e "\\bash echo 'Hello World'\n\\bash pwd\n\\bash ls | head -3\n\\exit" | ./bin/neuro shell