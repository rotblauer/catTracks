#!/usr/bin/env bash

# At 20240221:20:00 the ~/track.areteh.co/ directory was nuked. Nothing in it except favicons.
# No cat tracks executable, no cat tracks data.
# This caused track pushes to break because the 'linux.trackermain' process there
# is the point of entry for all cat trackers.
# So until we move to a reliable server, this is what I did to start it working again.

set -e
set -x

go build -o /tmp/linux.trackermain .
rsync -avz /tmp/linux.trackermain freya:~/track.areteh.co/
rsync -avz ./kickstart freya:~/track.areteh.co/

ssh freya 'cd track.areteh.co; pkill trackermain; ./kickstart'

# Make sure `kickstart` script is executable on Freya.
# She has a cron (trackalivecheckerkeeper) that will regularly restart
# this process if not already running. Another awesome workaround
# for living in Freya's fascist world.
