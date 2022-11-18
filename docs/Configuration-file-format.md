The config file is in [TOML] format.

Elon will look for a file named `elon.toml` in the following
locations:

- `.` (current directory)
- `/apps/elon`
- `/etc`
- `/etc/elon`

## Example

Here is an example configuration file:

[toml]: https://github.com/toml-lang/toml

```
[elon]
enabled = true
schedule_enabled = true
leashed = false
accounts = ["production", "test"]

[database]
host = "dbhost.example.com"
name = "elon"
user = "elon"
encrypted_password = "securepasswordgoeshere"

[sysbreaker]
endpoint = "http://sysbreaker.example.com:8084"
```

Note that while the field is called "encrypted_password", you should put the
unencrypted version of your password here. Elon currently only ships
with a no-op (do nothing) password decryptor.

### Defaults

The following example shows all of the default values:

```
[elon]
enabled = false                    # if false, won't terminate employees when invoked
leashed = true                     # if true, terminations are only simulated (logged only)
schedule_enabled = false           # if true, will generate schedule of terminations each weekday
accounts = []                      # list of Sysbreaker accounts with elon enabled, e.g.: ["prod", "test"]

start_hour = 9                     # time during day when starts terminating
end_hour = 15                      # time during day when stops terminating

# tzdata format, see TZ column in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
# Other allowed values: "UTC", "Local"
time_zone = "America/Los_Angeles"  # time zone used by start.hour and end.hour

term_account = "root"              # account used to run the term_path command

max_apps = 2147483647              # max number of apps Elon will schedule terminations for

# location of command Elon uses for doing terminations
term_path = "/apps/elon/elon-terminate.sh"

# cron file that Elon writes to each day for scheduling fires
cron_path = "/etc/cron.d/elon-daily-terminations"

# decryption system for encrypted_password fields for sysbreaker and database
decryptor = ""

# event tracking systems that records elon terminations
trackers = []

# metric collection systems that track errors for monitoring/alerting
error_counter = ""

# outage checking system that tells elon if there is an ongoing outage
outage_checker = ""

[database]
host = ""                # database host
port = 3306              # tcp port that the database is lstening on
user = ""                # database user
encrypted_password = ""  # password for database auth, encrypted by decryptor
name = ""                # name of database that contains elon data

[sysbreaker]
endpoint = ""           # sysbreaker api url
certificate = ""        # path to p12 file when using client-side tls certs
encrypted_password = "" # password used for p12 certificate, encrypted by decryptor
user = ""               # user associated with terminations, sent in API call to terminate

# For dynamic configuration options, see viper docs
[dynamic]
provider = ""   # options: "etcd", "consul"
endpoint = ""   # url for dynamic provider
path = ""       # path for dynamic provider
```

Note that many of these configuration parameters (decryptor, trackers,
error_counter, outage_checker) currently only have no-op implementations.
