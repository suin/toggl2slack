# Toggl2Slack

Post Toggl timer entries to Slack channel.

## Installation

Make config file:

```
mkdir ~/.toggl2slack
touch ~/.toggl2slack/config.json
```

Copy and paste API tokens

```json:config.json
{
	"SlackToken": "xoxp-00000000000000000000000000000000",
	"SlackChannel": "#times_suin",
  "TogglToken": "00000000000000000000000000000000"
}
```

## Build App

```
make bundle
```
