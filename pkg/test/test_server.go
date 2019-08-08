package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sanguohot/log"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

var (
	dir = os.Getenv("serve_dir")
	token = os.Getenv("token")
	redirect_to = os.Getenv("redirect_to")
	host = "0.0.0.0"
	port = 4200
)

func middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "0")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Next()
	}
}

func init()  {
	if dir == "" || token == "" || redirect_to == "" {
		log.Logger.Fatal("invalid param",
			zap.String("dir", dir),
			zap.String("token", token),
			zap.String("redirect_to", redirect_to),
			)
	}
}

func noRouteHandler(c *gin.Context) {
	p := filepath.Join(dir, c.Request.RequestURI)
	info, err := os.Stat(p)
	if err != nil {
		log.Logger.Error(err.Error())
	}
	if info.IsDir() {
		log.Logger.Error(fmt.Sprintf("dir %s not supported", p))
	}
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Logger.Error(err.Error())
	}
	if strings.HasPrefix(p, "/api") {
		c.Request.Header.Add("token", token)
		c.Redirect(http.StatusMovedPermanently, redirect_to)
		c.Abort()
	} else if strings.HasSuffix(p, ".json") {
		c.Data(http.StatusOK, "application/json", data)
	} else {
		c.Data(http.StatusOK, http.DetectContentType(data), data)
	}
}

func main() {
	//http.Handle("/", http.FileServer(http.Dir("e:/test/")))
	//http.ListenAndServe(":4200", nil)
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Recovery())
	// 默认设置logger，但启用logger会导致吞吐量大幅度降低
	if os.Getenv("GIN_LOG") != "off" {
		r.Use(gin.Logger())
	}
	r.Use(middleware())
	r.MaxMultipartMemory = 10 << 20 // 10 MB
	r.NoRoute(noRouteHandler)
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Logger.Fatal(err.Error())
		}
	}()
	log.Sugar.Infof("[http] listening => %s, serv => %s", server.Addr, dir)
	// apiserver发生错误后延时五秒钟，优雅关闭
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Logger.Fatal(err.Error())
	}
	log.Sugar.Infof("stop server => %s, serv => %s", server.Addr, dir)
}