#!/bin/sh

# Run planguard and capture output and exit code
/usr/local/bin/planguard "$@" > /tmp/planguard-output.txt
EXIT_CODE=$?

# Read the output
OUTPUT=$(cat /tmp/planguard-output.txt)

# If format is SARIF, write to file in workspace
if echo "$@" | grep -q "\-format sarif"; then
    echo "$OUTPUT" > /github/workspace/planguard-results.sarif
fi

# Always print output to stdout
echo "$OUTPUT"

# Exit with planguard's exit code
exit $EXIT_CODE
