# Summary

Simple slack triage bot that fetches all messages in a channel over a certain timespan and confirms that they have been responded to. Bot looks for messages that are missing emojis. This is designed to be a quick "someone has looked at this message" :eyes: and "someone has triaged this message completely" :white_check_mark:

## Configuration

The default configuration for this bot is

```
var ackEmojis = []string {"eyes"} #list of emojis to count for "acknowledged"
var doneEmojis = []string {"white_check_mark"} #list of emojis to count for "done"

var scrollbackDays int = -1 #days to look back. -1 = 24 hours, -2 = 48 hours

var reportingChannelNames = []string {"channel-name-1"} #list of channels to look in
var slackAPIKeyPath = "/path/to/ssm/secret" #location of the slack api key
```

# Setup

1) AWS Lambda function already created with access to read SSM parameter store 

2) Event bridge cron configured to run lambda on a routine basis. Suggested deployment is at the start of your working hours 

3) Slack App with an authorization token stored in SSM parameter store granting the below permissions

```
#(Required) - Access to read and write to channels
reactions:read
chat:write

#(Optional) Only required if reporting in public channels
channels:history
channels:read

#(Optional) Only required if reporting on private channels
groups:history
groups:read
```

4) Add the slack app into the channels you want to monitor

5) Update the conf with the channels you want to report on and let slacker tell you or your team missed a message

# Deploying

1) Login to AWS CLI using your preferred method. Update the Makefile to have your S3 bucket and lambda function name

2) Run the below commands 
```make build
make deploy
make publish```