# process-watcher

This is a simple process watcher.  I use this to create a virtual webcam, and start OBS.  This app ensures that 
the webcam is always running, and that OBS is always running.  There is an example commands.yaml file which shows how to 
setup commands for this to watch.  ctrl+c sends a SIGTERM to the processes and gracefully exits.

I dont really intend on fixing issues, maintaining, or doing anything else with this outside of personal use.

## Building

go build -o ~/.bin/process-watcher *.go