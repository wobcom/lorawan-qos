package run

import (
	"context"
	"net/http"
	"net/url"
	"network-qos/internal/app"
	"network-qos/internal/config"
	"network-qos/internal/storage"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	log.SetLevel(log.DebugLevel)
	log.Info("Starting Network-QoS-Service")

	err := storage.Setup(config.C)
	if err != nil {
		panic(err)
	}
	log.Info("Storage initialized")

	uri, err := url.Parse(config.C.Integration.DSN)
	if err != nil {
		panic(err)
	}

	topic := uri.Path[1:len(uri.Path)]
	log.Debugf("topic %s \n", topic)

	go app.Listen(uri, topic)

	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), config.C.General.ShutdownTimeout)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
	return nil
}
