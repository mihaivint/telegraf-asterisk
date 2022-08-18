# Asterisk Input Plugin

This plugin gather basic metrics from asterisk 

### Configuration:
```toml
[[inputs.asterisk]]
   # Asterisk socket
   socket = "/var/run/asterisk/asterisk.ctl"
   # only used for tagging
   nodeid = ""
```
### Tags
nodeid

### Metrics output example: 
``` toml
active_calls=0i
last_reload=1060i
sip_monitored_offline=0i
sip_monitored_online=0i
sip_peers=2i
sip_unmonitored_offline=0i
sip_unmonitored_online=0i
system_uptime=1060i
total_calls=0i 
```
