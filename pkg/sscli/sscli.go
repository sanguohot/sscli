package sscli

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sanguohot/log"
	"github.com/sanguohot/sscli/pkg/common/file"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

var (
	serveTypeApi = "api"
	serveTypeDir  = "dir"
	headerSplit = ":"
	headerValSplit = ";"
)

type item struct {
	ty    string
	prefix string
	target string
	header map[string][]string
}

type Sscli struct {
	port  int
	host  string
	tys []string
	paths []string
	targets []string
	items []item
	hs []string
}

func New(port int, host string, tys, paths, targets, hs []string) *Sscli {
	return &Sscli{
		port: port,
		host: host,
		tys: tys,
		paths: paths,
		targets: targets,
		hs: hs,
	}
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

func (s *Sscli) checkAndFormatParams() {
	if len(s.tys) == 0 {
		log.Logger.Fatal("len(s.tys) == 0")
	}
	if len(s.tys) != len(s.paths) {
		log.Logger.Fatal("len(s.tys) != len(s.paths)")
	}
	if len(s.tys) != len(s.targets) {
		log.Logger.Fatal("len(s.tys) != len(s.targets)")
	}
	pathm := make(map[string]bool)
	var k = 0
	for i, t := range s.tys {
		switch t {
		case serveTypeDir, serveTypeApi:
			it := item{
				ty: t,
				prefix: s.paths[i],
				target: s.targets[i],
			}
			if t == serveTypeApi && k < len(s.hs) && strings.Index(s.hs[k], headerSplit) >= 0 {
				hl := strings.Split(s.hs[k], headerSplit)
				if len(hl) != 2 {
					log.Logger.Fatal("invalid header", zap.String("header", s.hs[k]))
				}
				headerKey := hl[0]
				headerVal := strings.Split(hl[1], headerValSplit)
				it.header = map[string][]string{headerKey: headerVal}
				k++
			}
			if t == serveTypeDir {
				if err := checkDir(s.targets[i]); err != nil {
					log.Logger.Fatal(err.Error(), zap.String("target", s.targets[i]))
				}
			}
			s.items = append(s.items, it)

		default:
			log.Logger.Fatal(fmt.Sprintf("param 'type' can only be %s or %s.", serveTypeApi, serveTypeDir),
				zap.String("type", t))
			break
		}
		if _, ok := pathm[s.paths[i]]; ok {
			log.Logger.Fatal("path already exist", zap.String("path", s.paths[i]))
		}
		pathm[s.paths[i]] = true
	}
}

func (it *item) serveReverse(c *gin.Context) {
	director := func(req *http.Request) {
		req = c.Request
		req.URL.Scheme = "http"
		req.URL.Host = it.target
		//req.Header["my-header"] = []string{r.Header.Get("my-header")}
		//// Golang camelcases headers
		//delete(req.Header, "My-Header")
		for k, v := range it.header {
			if len(req.Header[k]) == 0 {
				req.Header[k] = []string{}
			}
			for _, vv := range v {
				req.Header[k] = append(req.Header[k], vv)
			}
		}
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (it *item) serveDir(c *gin.Context) {
	var (
		err  error
		data []byte
	)
	realPath := filepath.Join(it.target, c.Request.RequestURI[len(it.prefix):])
	if !file.FilePathExist(realPath) {
		err = fmt.Errorf("uri '%s', file path '%s' not found", c.Request.RequestURI, realPath)
		goto end
	}
	if bl, err := file.FileIsDir(realPath); err != nil {
		goto end
	} else if bl {
		err = fmt.Errorf("realPath '%s' is a diretory", realPath)
		goto end
	}
	data, err = ioutil.ReadFile(realPath)
end:
	if err != nil {
		log.Sugar.Error(err.Error(), zap.String("realPath", realPath))
		c.JSON(http.StatusNotFound, err.Error())
	} else {
		c.Data(http.StatusOK, http.DetectContentType(data), data)
	}
}

func (s *Sscli) noRouteHandler(c *gin.Context) {
	for _, it := range s.items {
		if strings.Contains(c.Request.RequestURI, it.prefix) {
			switch it.ty {
			case serveTypeApi:
				it.serveReverse(c)
			case serveTypeDir:
				it.serveDir(c)
			}
			break
		}
	}
}

func (s *Sscli) startServer() {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware())
	// 默认设置logger，但启用logger会导致吞吐量大幅度降低
	if os.Getenv("GIN_LOG") != "off" {
		r.Use(gin.Logger())
	}
	r.MaxMultipartMemory = 10 << 20 // 10 MB
	r.NoRoute(s.noRouteHandler)
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
	log.Sugar.Infof("[http] listening => %s, serve => %v", server.Addr, s.items)
	// apiserver发生错误后延时五秒钟，优雅关闭
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Logger.Fatal(err.Error())
	}
	log.Sugar.Infof("stop server => %s, serve => %v", server.Addr, s.items)
}

func (s *Sscli) Serve() {
	s.checkAndFormatParams()
	s.startServer()
}
