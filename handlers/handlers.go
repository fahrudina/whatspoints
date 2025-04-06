package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wa-serv/database"
	"github.com/wa-serv/processor"
	"github.com/wa-serv/s3uploader"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func HandleMessageEvent(v *events.Message, db *sql.DB, client *whatsmeow.Client) {
	var msgText string
	if v.Message.GetExtendedTextMessage().GetText() != "" {
		msgText = v.Message.GetExtendedTextMessage().GetText()
	} else {
		msgText = v.Message.GetConversation()
	}

	fmt.Printf("Received message from %s: %s\n", v.Info.Sender.String(), msgText)

	if v.Message.GetImageMessage() != nil {
		handleMediaMessage(v, db, client)
	} else {
		err := processor.ProcessRegistration(client, db, msgText, v.Info.Sender.String())
		if err != nil {
			fmt.Printf("Registration processing error: %v\n", err)
		}

		if msgText == "ping" {
			replyToMessage(v, client)
		} else if msgText == "help" {
			sendHelpMessage(v, client)
		}
	}
}

func handleMediaMessage(evt *events.Message, db *sql.DB, client *whatsmeow.Client) {
	imageMessage := evt.Message.GetImageMessage()
	if imageMessage != nil {
		fmt.Printf("Received an image message from %s\n", evt.Info.Sender.String())

		data, err := client.Download(imageMessage)
		if err != nil {
			fmt.Printf("Failed to download image: %v\n", err)
			return
		}

		memberID, err := database.GetMemberIDByPhoneNumber(db, evt.Info.Sender.String())
		if err != nil {
			fmt.Printf("Failed to retrieve member ID: %v\n", err)
			return
		}

		imageURL, err := s3uploader.UploadToS3(data)
		if err != nil {
			fmt.Printf("Failed to upload image to S3: %v\n", err)
			return
		}

		err = database.SaveImageURL(db, memberID, imageURL)
		if err != nil {
			fmt.Printf("Failed to save image URL to database: %v\n", err)
			return
		}

		msg := &waProto.Message{
			Conversation: proto.String("Image received and saved successfully."),
		}
		_, err = client.SendMessage(context.Background(), evt.Info.Sender, msg)
		if err != nil {
			fmt.Printf("Error sending acknowledgment: %v\n", err)
		}
	}
}

func replyToMessage(evt *events.Message, client *whatsmeow.Client) {
	msg := &waProto.Message{
		Conversation: proto.String("pong"),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func sendHelpMessage(evt *events.Message, client *whatsmeow.Client) {
	helpText := `Available commands:
- ping: Bot responds with "pong"
- help: Shows this help message`

	msg := &waProto.Message{
		Conversation: proto.String(helpText),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending help message: %v\n", err)
	}
}
