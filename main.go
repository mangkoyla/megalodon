package main

import (
	"sync"

	database "github.com/FoolVPN-ID/Megalodon/db"
	logger "github.com/FoolVPN-ID/Megalodon/log"
	"github.com/FoolVPN-ID/Megalodon/provider"
	"github.com/FoolVPN-ID/Megalodon/sandbox"
	"github.com/joho/godotenv"
)

func main() {
	logger := logger.MakeLogger()
	godotenv.Load()
	db := database.MakeDatabase()
	prov := provider.MakeSubProvider()
	sb := sandbox.MakeSandbox()

	// Nodes gathering
	logger.Info("Gathering nodes...")
	prov.GatherSubFile()
	prov.GatherNodes()

	// Goroutine goes here 💪🏻
	var (
		wg    = sync.WaitGroup{}
		queue = make(chan struct{}, 100)
	)

	logger.Info("Processing...")
	for i, node := range prov.Nodes {
		wg.Add(1)
		queue <- struct{}{}

		go func() {
			defer func() {
				<-queue
				wg.Done()
			}()

			sb.TestConfig(node, i, len(prov.Nodes))
		}()
	}

	// Wait for all concurrency to be done
	logger.Info("Waiting for goroutines...")
	wg.Wait()

	// Save results to database
	logger.Info("Saving results to database...")
	if err := db.Save(sb.Results); err == nil {
		// Sync local to remote
		db.SyncAndClose()
	}

}
