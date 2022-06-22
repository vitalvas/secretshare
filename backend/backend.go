package backend

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v3"
)

type Backend struct {
	auth  *Auth
	db    *badger.DB
	crypt Crypt
}

func Execute() {
	app := Backend{
		auth: NewAuth(),
		db:   NewDB(),
	}

	app.crypt.PSK = os.Getenv("APP_SECRET_PSK")

	go app.runGC()

	defer app.db.Close()

	httpServer := &http.Server{
		Addr:        ":8000",
		Handler:     app.GetRouter(),
		IdleTimeout: time.Minute,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-notifyCtx.Done()

	log.Println("shutdown")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), httpServer.IdleTimeout)
	defer cancel()

	if err := httpServer.Shutdown(timeoutCtx); err != nil {
		log.Fatal(err)
	}
}

func NewDB() *badger.DB {
	dbOpts := badger.DefaultOptions("./db")
	dbOpts = dbOpts.WithValueLogFileSize(128 << 20) // 128MB
	dbOpts = dbOpts.WithIndexCacheSize(128 << 20)   // 128MB
	dbOpts = dbOpts.WithBaseTableSize(8 << 20)      // 8MB
	dbOpts = dbOpts.WithCompactL0OnClose(true)

	if value, ok := os.LookupEnv("APP_ENCRYPTION_KEY"); ok {
		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			log.Fatal(err)
		}

		if len(data) != 32 {
			log.Fatal("Encryption key length must be 32 bytes")
		}

		dbOpts = dbOpts.WithEncryptionKey(data)
		dbOpts = dbOpts.WithEncryptionKeyRotationDuration(7 * 24 * time.Hour) // 7 days
	}

	db, err := badger.Open(dbOpts)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func (backend *Backend) runGC() {
	for {
		time.Sleep(10 * time.Minute)
		if err := backend.db.RunValueLogGC(0.7); err != nil {
			if err != badger.ErrNoRewrite {
				log.Println("error db gc:", err)
			}
		}
	}
}
