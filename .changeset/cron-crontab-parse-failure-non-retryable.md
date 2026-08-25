---
"chainlink": patch
---

#bugfix Pin `capability-cron` to the release that fixes a scheduler goroutine leak on every rejected cron schedule, and classify gocron's `crontab parse failure` message as a permanent activation error so the workflow syncer stops retrying an activation whose schedule can never parse.
