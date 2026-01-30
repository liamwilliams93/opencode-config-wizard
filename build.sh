#!/bin/bash
echo "Building OpenCode Config Wizard for Linux..."
go build -o opencode-config-wizard .
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi
echo "Build successful: opencode-config-wizard"
chmod +x opencode-config-wizard
