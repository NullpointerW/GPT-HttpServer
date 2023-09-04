package main

import (
	"context"
	"github.com/gin-gonic/gin"
	_ "gpt3.5/cache"
	"gpt3.5/cfg"
	gptHttp "gpt3.5/http"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	log.Printf("http service listens on port [%d]", cfg.Cfg.Port)
	ginSrv(strconv.Itoa(cfg.Cfg.Port))
	// mux := http.NewServeMux()
	// mux.Handle("/v1/chat/do", http.HandlerFunc(gptHttp.Do))
	// mux.Handle("/cfg/modifyKey", http.HandlerFunc(gptHttp.SwitchApikey))
	// log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Cfg.Port), mux))
}

func ginSrv(port string) {
	gin.SetMode(gin.ReleaseMode)
	router := gptHttp.SetupRouter()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
