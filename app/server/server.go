// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/log"

	api "github.com/mendersoftware/reporting/api/http"
	"github.com/mendersoftware/reporting/app/reporting"
	dconfig "github.com/mendersoftware/reporting/config"
	"github.com/mendersoftware/reporting/store"
)

func init() {
	if mode := os.Getenv(gin.EnvGinMode); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

// InitAndRun initializes the server and runs it
func InitAndRun(conf config.Reader, store store.Store) error {
	ctx := context.Background()

	l := log.FromContext(ctx)

	reporting := reporting.NewApp(store)

	var listen = conf.GetString(dconfig.SettingListen)
	var router = api.NewRouter(reporting)
	srv := &http.Server{
		Addr:    listen,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)
	<-quit

	l.Info("Shutdown Server ...")

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxWithTimeout); err != nil {
		l.Fatal("Server Shutdown: ", err)
	}

	return nil
}
