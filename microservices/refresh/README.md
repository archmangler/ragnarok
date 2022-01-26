# Service watchdog

A small service to monitor producers and restart the producers in case they get stuck 

- watch for inactivity from all producers
- restart the producer deployment ("rollout restart" with kubectl)
- check restart happened properly, retry on fail

