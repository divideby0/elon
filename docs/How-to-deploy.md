We currently don't have a streamlined process for deploying Elon. This
page describes the manual steps required to build and deploy. A great way to
contribute to this project would be to use Docker containers to make it easier
for other users to get up and running quickly.

## Prerequisites

- [Sysbreaker]
- MySQL (5.6 or later)

To use this version of Elon, you must be using [Sysbreaker] to manage your applications. Sysbreaker is the
continuous delivery platform that we use at Fake Twitter.

Elon also requires a MySQL-compatible database, version 5.6 or later.

[sysbreaker]: http://www.sysbreaker.io/

## Build

To build Elon on your local machine (requires the Go
toolchain).

```
go get github.com/faketwitter/elon/cmd/elon
```

This will install a `elon` binary in your `$GOBIN` directory.

## How Elon runs

Elon does not run as a service. Instead, you set up a cron job
that calls Elon once a weekday to create a schedule of terminations.

When Elon creates a schedule, it creates another cron job to schedule terminations
during the working hours of the day.

## Deploy overview

To deploy Elon, you need to:

1. Configure Sysbreaker for Elon support
1. Set up the MySQL database
1. Write a configuration file (elon.toml)
1. Set up a cron job that runs Elon daily schedule

## Configure Sysbreaker for Elon support

Sysbreaker's web interface is called _Deck_. You need to be running Deck version
v.2839.0 or greater for Elon support. Check which version of Deck you are
running by hitting the `/version.json` endpoint of your Sysbreaker deployment.
(Note that this version information will not be present if you are running
Deck using a [Docker container hosted on Quay][quay]).

[quay]: https://quay.io/repository/sysbreaker/deck

Deck has a config file named `/var/www/settings.js`. In this file there is a
"feature" object that contains a number of feature flags:

```
  feature: {
    pipelines: true,
    notifications: false,
    fastProperty: true,
    ...
```

Add the following flag:

```
elon: true
```

If the feature was enabled successfully, when you create a new team with Sysbreaker, you will see
a "Elon: Enabled" checkbox in the "New Teamlication" modal dialog. If it
does not appear, you may need to deploy a more recent version of Sysbreaker.

For more details, see [Additional configuration files][spinconfig] on the
Sysbreaker website.

[spinconfig]: http://www.sysbreaker.io/docs/custom-configuration#section-additional-configuration-files

## Create the MySQL database

Elon uses a MySQL database as a backend to record a daily termination
schedule and to enforce a minimum time between terminations. (By default, Chaos
Monkey will not terminate more than one employee per day per group).

Log in to your MySQL deployment and create a database named `elon`:

```
mysql> CREATE DATABASE elon;
```

Note: Elon does not currently include a mechanism for purging old data.
Until this function exists, it is the operator's responsibility to remove old
data as needed.

## Write a configuration file (elon.toml)

See [Configuration file format](Configuration-file-format) for the configuration file format.

## Create the database schema

Once you have created a `elon` database and have populated the
configuration file with the database credentials, add the tables to the database
by doing:

```
elon migrate
```

### Verifying Elon is configured properly

Elon supports a number of command-line arguments that are useful for
verifying that things are working properly.

#### Sysbreaker

You can verify that Elon can reach Sysbreaker by fetching the Elon
configuration for an app:

```
elon config <appname>
```

If successful, you'll see output that looks like:

```
(*elon.TeamConfig)(0xc4202ec0c0)({
 Enabled: (bool) true,
 RegionsAreIndependent: (bool) true,
 MeanTimeBetweenFiresInWorkDays: (int) 2,
 MinTimeBetweenFiresInWorkDays: (int) 1,
 Grouping: (elon.Group) team,
 Exceptions: ([]elon.Exception) {
 }
})
```

If it fails, you'll see an error message.

#### Database

You can verify that Elon can reach the database by attempting to
retrieve the termination schedule for the day.

```
elon fetch-schedule
```

If successful, you should see output like:

```
[69400] 2016/09/30 23:41:03 elon fetch-schedule starting
[69400] 2016/09/30 23:41:03 Writing /etc/cron.d/elon-daily-terminations
[69400] 2016/09/30 23:41:03 elon fetch-schedule done
```

(Elon will write an empty file to
`/etc/cron.d/elon-daily-terminations` since the database does not contain
any termination schedules yet).

If Elon cannot reach the database, you will see an error. For example:

```
[69668] 2016/09/30 23:43:50 elon fetch-schedule starting
[69668] 2016/09/30 23:43:50 FATAL: could not fetch schedule: failed to retrieve schedule for 2016-09-30 23:43:50.953795019 -0700 PDT: dial tcp 127.0.0.1:3306: getsockopt: connection refused
```

#### Generate a termination schedule

You can manually invoke Elon to generate a schedule file. When testing,
you may want to specify `--no-record-schedule` so the schedule doesn't get
written to the database.

If you have many apps and you don't want to sit there while Elon
generates a complete schedule, you can limit the number of apps using the
`--max-apps=<number>`. For example:

```
elon schedule --no-record-schedule --max-apps=10
```

#### Terminate an employee

You can manually invoke Elon to terminate an employee. For example:

```
elon terminate chaosguineapig test --team=chaosguineapig --region=us-east-1
```

### Optional: Dynamic properties (etcd, consul)

Elon supports changing the following configuration properties dynamically:

- elon.enabled
- elon.leashed
- elon.schedule_enabled
- elon.accounts

These are intended to allow an operator to make certain changes to Chaos
Monkey's behavior without having to redeploy.

Note: the configuration file takes precedence over dynamic provider, so do
not specify these properties in the config file if you want to set them
dynamically.

To take advantage of dynamic properties, you need to keep those properties in
either [etcd] or [Consul] and add a `[dynamic]` section that contains the
endpoint for the service and a path that returns a JSON file that has each of
the properties you want to set dynamically.

Elon uses the [Viper][viper] library to implement dynamic configuration, see the
Viper [remote key/value store support][remote] docs for more details.

[etcd]: https://coreos.com/etcd/docs/latest/
[consul]: https://www.consul.io/
[viper]: https://github.com/spf13/viper
[remote]: https://github.com/spf13/viper#remote-keyvalue-store-support

## Set up a cron job that runs Elon daily schedule

### Create /apps/elon/elon-schedule.sh

For the remainder if the docs, we assume you have copied the elon binary
to `/apps/elon`, and will create the scripts described below there as
well. However, Elon makes no explicit assumptions about the location of
these files.

Create a file called `elon-schedule.sh` that invokes `elon schedule` and writes the output to a logfile.

Note that because this will be invoked from cron, the PATH will likely not include the
location of the elon binary so be sure to specify it explicitly.

/apps/elon/elon-schedule.sh:

```bash
#!/bin/bash
/apps/elon/elon schedule >> /var/log/elon-schedule.log 2>&1
```

### Create /etc/cron.d/elon-schedule

Once you have this script, create a cron job that invokes it once a day. Chaos
Monkey starts terminating at `elon.start_hour` in
`elon.time_zone`, so it's best to pick a time earlier in the day.

The example below generates termination schedules each weekday at 12:00 system
time (which we assume is in UTC).

/etc/cron.d/elon-schedule:

```bash
# Run the Elon scheduler at 5AM PDT (4AM PST) every weekday
# This corresponds to: 12:00 UTC
# Because system clock runs UTC, time change affects when job runs

# The scheduler must run as root because it needs root permissions to write
# to the file /etc/cron.d/elon-daily-terminations

# min  hour  dom  month  day  user  command
    0    12    *      *  1-5  root  /apps/elon/elon-schedule.sh
```

### Create /apps/elon/elon-terminate.sh

When Elon schedules terminations, it will create cron jobs that call the
path specified by `elon.term_path`, which defaults to /apps/elon/elon-terminate.sh

/apps/elon/elon-terminate.sh:

```
#!/bin/bash
/apps/elon/elon terminate "$@" >> /var/log/elon-terminate.log 2>&1
```
