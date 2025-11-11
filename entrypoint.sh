#!/bin/sh

# Run planguard and capture output and exit code (stderr + stdout)
/usr/local/bin/planguard "$@" > /tmp/planguard-output.txt 2>&1
EXIT_CODE=$?

# Read the output
OUTPUT=$(cat /tmp/planguard-output.txt)

# Set GitHub Action outputs
if [ -n "$GITHUB_OUTPUT" ]; then
    # Set passed output
    if [ "$EXIT_CODE" -eq 0 ]; then
        echo "passed=true" >> "$GITHUB_OUTPUT"
    else
        echo "passed=false" >> "$GITHUB_OUTPUT"
    fi

    # Set violations output (only for JSON format)
    if echo "$@" | grep -qE "format.*(json|sarif)"; then
        # Escape and set violations as multiline output
        echo "violations<<EOF" >> "$GITHUB_OUTPUT"
        echo "$OUTPUT" >> "$GITHUB_OUTPUT"
        echo "EOF" >> "$GITHUB_OUTPUT"
    fi
fi

# If format is SARIF, write to file in workspace (improved detection)
if echo "$@" | grep -qE "format.*sarif"; then
    echo "$OUTPUT" > /github/workspace/planguard-results.sarif
fi

# Always print output to stdout
echo "$OUTPUT"

# Exit with planguard's exit code
exit $EXIT_CODE
