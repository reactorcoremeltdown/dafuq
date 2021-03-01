# DAFUQ

![](https://ci.rcmd.space/badge/dafuq.svg)

...so, uhm, 'dafuq just happened?

### TABLE OF CONTENTS

+ [Description]()
+ [Motivation]()
+ [So, how do I write a config?]()
+ [And now, how do I launch it?]()
+ [That looks great! But how do I check live status?]()
+ [Building]()
+ [Roadmap]()

### Description

Simply put, `dafuq` is an answering machine. It can help you resolve questions like "DAFUQ is happening with my server?" by running scripts and checking their exit codes for changes.

### Motivation

Yes, there are lots of similar projects, take various `nagios-core` implementations for example, or `monit`. However, there are certain problems that I find rather discouraging when setting up monitoring systems:

+ Configs are bloated and too expensive to maintain
+ Sometimes packages are getting expelled from official repositories of my favorite operating system

The way I like it includes nothing else other than writing a small number of really simple configs, throwing a binary on a machine, then just launching it. Simplicity is what I aim for!

### So, how do I write a config?

By default, `dafuq` reads its configs stored at `/etc/dafuq`. Configs are written in a dead simple INI format.

The main config file is called `config.ini` and has a following structure:

```
[main]
configs = /etc/dafuq/configs
plugins = /etc/dafuq/plugins
notifiers = /etc/dafuq/notifiers
address = 127.0.0.1
port = 8881
stateFile=/var/lib/dafuq/dafuq.state
```

Every field is explained below:

+ `configs` — where to find configs of monitoring checks
+ `plugins` — where to find scripts that perform actual checks (shell scripts, exit codes are responsible for changing the status of checks)
+ `notifiers` — where to find executables to throw alerts (PagerDuty, Slack, you name it)
+ `address` — an IP to bind to, usually it's just localhost
+ `port` — a TCP port to listen on
+ `stateFile` — where to dump a JSON state of all checks on every status change to survive restarts


That's it. Main app config is ready to use!

Now, to the monitoring check configs. A typical check config is listed below:

```
[config]
name = provisioning
description = Provisioning file
plugin = check_file.sh
argument = /etc/default/earlystageconfigs
interval = 15s
notify = gotify.sh
```

Each field is explained below:

+ `name` — pretty self-explanatory, isn't it? Better to use lowercase with no whitespace characters, but I'm not stopping you here
+ `description` — put here anything that would help you to better understand 'dafuq has happened @ 3:00 AM if your service goes belly up
+ `plugin` — which script from `plugins` directory to use to perform an actual check. **This needs to be executable**.
+ `argument` — which argument (`$1`) to pass to the script above
+ `interval` — at which interval to run this check. This field should support common shorthands for parsing time suffixes(`s`/`m`/`h`/`d`)
+ `notify` — which script to use to bug you when your check goes bonkers. **This needs to be executable as well**.


Notice: `dafuq` itself does not care at slightest about the actual state of matters, it only fires up a script from `notifiers` directory when the exit code of a plugin _changes_. Then, it's your job to define whether something is good or bad.

An example of a plugin script:

```
#!/usr/bin/env bash
source /etc/dafuq/plugins/okfail

if [[ -f $1 ]]; then
	ok "File $1 is in place." "$DESCRIPTION" "$ENVIRONMENT"
else
	fail "File $1 is not in place!" "$DESCRIPTION" "$ENVIRONMENT"
fi
```

For usability sake I also made a script up as an `okfail` "library":

```
function ok() {
    echo -n "OK - "$1
    exit 0
}

function warning() {
    echo -n "WARNING - "$1
    exit 1
}

function fail() {
    echo -n "CRITICAL - "$1
    exit 2
}
```

And now, we need to throw an alert when status changes:

```
#!/usr/bin/env bash
echo ${MESSAGE} ${STATUS} ${NAME} ${DESCRIPTION} >> /tmp/monitoring.log
```

The variables above are environment variables that `dafuq` sets for every notifier script.

### And now, how do I launch it?

Use systemd, or whatever you like.

For systemd, this is what the config looks like:

```
[Unit]
Description=WTF just happened?

[Service]
Environment="CONFIG_PATH=/etc/dafuq/config.ini"
ExecStart=/usr/local/bin/dafuq

[Install]
WantedBy=multi-user.target
```

Using `CONFIG_PATH` environment variable you can change the path of the main config to scan!

### That looks great! But how do I check live status?

Use `/` HTTP route to fetch the status in JSON format, for example `curl -s http://localhost:8881/`

### Building

There's nothing difficult, just run `go get && go build`. Requires Go 1.12 at least (due to native support of exit codes)

The Makefile here is for continuous integration.

### Roadmap

I have no plans for further development of `dafuq`, except of fixing possible bugs and doing occasional maintenance releases. If you have any concerns, please let me know in the issues!

DAFUQ is licensed under GPL version 3, so please contribute back your changes ;)
