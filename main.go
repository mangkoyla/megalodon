package main

import (
	"fmt"
	"sync"
	"time"

	database "github.com/FoolVPN-ID/Megalodon/db"
	logger "github.com/FoolVPN-ID/Megalodon/log"
	"github.com/FoolVPN-ID/Megalodon/provider"
	"github.com/FoolVPN-ID/Megalodon/sandbox"
	"github.com/FoolVPN-ID/Megalodon/telegram/bot"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	var (
		bot    = bot.MakeTGgBot()
		logger = logger.MakeLogger()
		db     = database.MakeDatabase()
		prov   = provider.MakeSubProvider()
		sb     = sandbox.MakeSandbox()
	)

	// Send notification to admin
	bot.SendTextToAdmin("Megalodon started!")

	// Nodes gathering
	logger.Info("Gathering nodes...")
	prov.GatherSubFile()
	prov.GatherNodes()

	// Goroutine goes here 💪🏻
	var (
		wg     = sync.WaitGroup{}
		queue  = make(chan struct{}, 200)
		isDone = false
	)

	// Report progress each minute
	go func() {
		for {
			time.Sleep(60 * time.Second)
			if isDone {
				break
			}

			bot.SendTextToAdmin(fmt.Sprintf("Account successfully tested: %d", len(sb.Results)))
		}
	}()

	logger.Info("Processing...")
	for i, rawConfig := range prov.Nodes {
		wg.Add(1)
		queue <- struct{}{}

		go func(node string, currentCount, maxCount int) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(fmt.Sprintf("Recover from panic: %v", err))
				}

				wg.Done()
				<-queue
			}()

			sb.TestConfig(node, currentCount, maxCount)
		}(rawConfig, i, len(prov.Nodes))
	}

	// Wait for all concurrency to be done
	logger.Info("Waiting for goroutines...")
	wg.Wait()

	// Stop reporting
	isDone = true

	// Save results to database
	logger.Info("Saving results to database...")
	if err := db.Save(sb.Results); err == nil {
		// Sync local to remote
		db.SyncAndClose()
	}

	bot.SendTextToAdmin("Megalodon finished!")
}
