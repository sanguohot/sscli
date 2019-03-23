package sscli

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sanguohot/filecp/pkg/common/file"
	"github.com/sanguohot/sscli/pkg/common/log"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Sscli struct {
	port  int
	host  string
	paths []string
	dirs  []string
}

func New(port int, host string, paths, dirs []string) *Sscli {
	return &Sscli{port: port, host: host, paths: paths, dirs: dirs}
}

func checkDir(dir string) error {
	if dir == "" {
		return errors.New("dir should not be empty string")
	}
	if !file.FilePathExist(dir) {
		log.Sugar.Infof("dir %s not found, create now...", dir)
		if err := file.EnsureDir(dir); err != nil {
			return err
		}
	} else if b, err := file.FileIsDir(dir); err != nil {
		return err
	} else if !b {
		return fmt.Errorf("file path %d exit, but require to be dir", dir)
	}
	return nil
}

func (s *Sscli) checkParams() {
	if len(s.paths) != len(s.dirs) {
		log.Logger.Fatal("param 'paths' and 'dirs' length should be equal.",
			zap.String("host", s.host), zap.Strings("dirs", s.dirs), zap.Strings("paths", s.paths))
	}
	if len(s.dirs) <= 0 || len(s.host) <= 0 || len(s.paths) <= 0 {
		log.Logger.Fatal("param 'host', 'dirs' or 'paths' invalid",
			zap.String("host", s.host), zap.Strings("dirs", s.dirs), zap.Strings("paths", s.paths))
	}
	if s.port < 1024 || s.port > 65535 {
		log.Logger.Fatal("port range should be 1024-65535", zap.Int("port", s.port))
	}
	pm := make(map[string]bool, len(s.paths))
	dm := make(map[string]bool, len(s.dirs))
	for k, _ := range s.dirs {
		if _, ok := pm[s.paths[k]]; ok {
			log.Logger.Fatal("found duplicate path", zap.String("path", s.paths[k]))
		}
		pm[s.paths[k]] = true
		if _, ok := dm[s.dirs[k]]; ok {
			log.Logger.Fatal("found duplicate dir", zap.String("dir", s.dirs[k]))
		}
		if err := checkDir(s.dirs[k]); err != nil {
			log.Logger.Fatal(err.Error(), zap.Int("port", s.port), zap.String("host", s.host), zap.String("dir", s.dirs[k]))
		}
		pm[s.dirs[k]] = true
	}
}

func (s *Sscli) startServer() {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Recovery())
	// 默认设置logger，但启用logger会导致吞吐量大幅度降低
	if os.Getenv("GIN_LOG") != "off" {
		r.Use(gin.Logger())
	}
	r.MaxMultipartMemory = 10 << 20 // 10 MB
	for k, v := range s.dirs {
		r.Static(s.paths[k], v)
	}
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", s.host, s.port),
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
	log.Sugar.Infof("[http] listening => %s, serve => paths:%v, dirs:%v", server.Addr, s.paths, s.dirs)
	// apiserver发生错误后延时五秒钟，优雅关闭
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Logger.Fatal(err.Error())
	}
	log.Sugar.Infof("stop server => %s, serve => paths:%v, dirs:%v", server.Addr, s.paths, s.dirs)
}

func (s *Sscli) Serve() {
	s.checkParams()
	s.startServer()
}
