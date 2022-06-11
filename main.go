package main

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

//globals

var ackEmojis = []string {"eyes"}
var doneEmojis = []string {"white_check_mark"}

var scrollbackDays int = -1

var reportingChannelNames = []string {"channel-name-1"}
var slackAPIKeyPath = "/path/to/ssm/secret"

var logOnly = false

func main() {
	lambda.Start(checkAlerts)
}

func checkAlerts() error {
	
	apiToken := getSSMSecret(slackAPIKeyPath)

	timeStamp := getLastTimeStamp(scrollbackDays)

	channels := getChannels(apiToken)

	for _, channel := range channels {

		fmt.Println("Checking messages in", channel.GroupConversation.Name)

		if ! stringListContainsString(reportingChannelNames, channel.GroupConversation.Name ){
			fmt.Println("Skipping channel", channel.GroupConversation.Name, " was not configured in channel list")
			continue
		}

		channelId := channel.GroupConversation.Conversation.ID

		messages := checkChannel(apiToken, channelId, timeStamp)

		missed, incomplete := checkMessages(messages)

		if missed > 0 || incomplete > 0 {
			warning := ":wave: There are messages that have not been triaged in the last 24 hours"
		
			if missed > 0 { 
				warning = warning + fmt.Sprint("\n\nThere were ", missed, " messages missed :eyes: ")
			}

			if incomplete > 0 {
				warning = warning + fmt.Sprint("\n\n There were ", incomplete, " messages with :eyes: but never completed with :white_check_mark: ")
			}

			fmt.Println(warning)
			if ! logOnly {
				err := sendSlackMessage(apiToken, channelId, warning)
				if err != nil {
					fmt.Printf("ERROR: Unable to write message to channel: %s\n", err)
				}
			}

		}
		time.Sleep(2 * time.Second) //little pause for slack rate limits
	}

	return nil
}

func checkMessages(messages []slack.Message) (int, int) {
	var missed int = 0
	var incomplete int = 0

	for _, message := range messages {

		if message.SubType == "bot_message" { //should filter out all other non alerts

			reactions := message.Reactions

			//if message was not marked complete
			//check if it was marked acknowledged
			//if not acknowledged then message was missed
			//if acknowledged then message is incomplete
			if ! checkReactions(reactions, doneEmojis) {
				if ! checkReactions(reactions, ackEmojis) {
					missed = missed + 1
				}else{
					incomplete = incomplete + 1
				}
			}
		}

	}
	return missed, incomplete
}

func stringListContainsString(list []string, matchString string) bool {
	for _, listEntry := range list {
		if matchString == listEntry {
			return true
		}
	}
	return false
}

func checkReactions(reactions []slack.ItemReaction, matchingReactions []string) bool {
	for _, reaction := range reactions {
		if stringListContainsString(matchingReactions, reaction.Name){
			return true
		}
	}
	return false
}

func checkChannel(apiToken string, channelId string, timeStamp string) []slack.Message{
	api := slack.New(apiToken)

	history, err := api.GetConversationHistory(&slack.GetConversationHistoryParameters {ChannelID: channelId, Oldest: timeStamp})

	if err != nil {
		fmt.Printf("ERROR: Unable to get message for channel. Err: %s\n", err)
	}
	
	return history.Messages
}

func sendSlackMessage(apiToken, channelId, message string) error {
	api := slack.New(apiToken)

	_, _, err := api.PostMessage(channelId, slack.MsgOptionText(message, false))

	return err
}

func getLastTimeStamp(days int) string {
	t := time.Now().AddDate(0, 0, days)
	timeStamp := t.Unix()
	return fmt.Sprint(timeStamp)
}

func getChannels(apiToken string) []slack.Channel {

	api := slack.New(apiToken)
	authTest, _ := api.AuthTest()
	userId := authTest.UserID //hacky way to get our user id

	publicConversations, _, err := api.GetConversationsForUser(&slack.GetConversationsForUserParameters {UserID: userId, Types: []string{"public_channel"}})

	if err != nil {
		fmt.Printf("WARNING: Unable to retrieve public conversations: %s\n", err)
	}

	privateConversations, _, err := api.GetConversationsForUser(&slack.GetConversationsForUserParameters {UserID: userId, Types: []string{"private_channel"}})

	if err != nil {
		fmt.Printf("WARNING: Unable to get private conversations. Check slack permissions allow for this action: %s\n", err)
	}

	return append(publicConversations, privateConversations...)
}

func getSSMSecret(secretName string) string {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:				aws.Config{Region: aws.String("us-east-1")},
		SharedConfigState: 	session.SharedConfigEnable,
	})
	if err != nil {
		panic(err)
	}

	ssmSvc := ssm.New(sess, aws.NewConfig().WithRegion("us-east-1"))
	param, err := ssmSvc.GetParameter(&ssm.GetParameterInput{
		Name: 				aws.String(secretName),
		WithDecryption:		aws.Bool(true),
	})

	if err != nil {
		panic(err)
	}

	value := *param.Parameter.Value
	return value
}
