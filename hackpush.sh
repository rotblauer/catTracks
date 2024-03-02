#!/usr/bin/env bash

set -e
set -x

# Build the new version of cattracks.
mkdir -p ./build/bin
go build -o ./build/bin/cattracks .

# Create a backup of the currently running cattracks version.
rsync -avz rotblauer.cattracks:/usr/local/bin/cattracks ./build/bin/cattracks.bak

# Copy the new version to the server and restart the service.
rsync -avz ./build/bin/cattracks rotblauer.cattracks:/usr/local/bin/cattracks
ssh rotblauer.cattracks "sudo systemctl restart cattracks.service"

echo "To revert:"
echo "rsync -avz ./build/bin/cattracks.bak rotblauer.cattracks:/usr/local/bin/cattracks && ssh rotblauer.cattracks 'sudo systemctl restart cattracks.service'"
