package bot

import (
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
	adminID, err := strconv.Atoi(os.Getenv("ADMIN_ID"))
	if err != nil {
		panic(err)
	}

	tgb := &tgBotStruct{
		token:   os.Getenv("BOT_TOKEN"),
		adminID: int64(adminID),
	}
	tgb.client = echotron.NewAPI(tgb.token)

	return tgb
}

func (tgb *tgBotStruct) SendTextFileToAdmin(filename, text, caption string) {
	file := echotron.NewInputFileBytes(filename, []byte(text))

	tgb.client.SendDocument(file, tgb.adminID, &echotron.DocumentOptions{Caption: caption})
}

func (tgb *tgBotStruct) SendTextToAdmin(text string) {
	tgb.client.SendMessage(text, tgb.adminID, nil)
}
