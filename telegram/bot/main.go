package bot

import (
	"log"
	"os"
	"strconv"

	"github.com/NicoNex/echotron/v3"
)

type tgBotStruct struct {
	token   string
	adminID int64
	client  echotron.API
}

func MakeTGgBot() *tgBotStruct {
	adminIDStr := os.Getenv("ADMIN_ID")
	adminID, err := strconv.Atoi(adminIDStr)
	if err != nil {
		log.Fatalf("Invalid ADMIN_ID: %v", err)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	tgb := &tgBotStruct{
		token:   token,
		adminID: int64(adminID),
	}
	tgb.client = echotron.NewAPI(tgb.token)

	return tgb
}

func (tgb *tgBotStruct) SendTextFileToAdmin(filename, text, caption string) {
	file := echotron.NewInputFileBytes(filename, []byte(text))

	_, err := tgb.client.SendDocument(file, tgb.adminID, &echotron.DocumentOptions{Caption: caption})
	if err != nil {
		log.Printf("Failed to send document: %v", err)
	}
}

func (tgb *tgBotStruct) SendTextToAdmin(text string) {
	_, err := tgb.client.SendMessage(text, tgb.adminID, nil)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
