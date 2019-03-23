# sscli
a command tool to serve multi static dir with gin
### install and use it
 ```
 $ git clone https://github.com/sanguohot/sscli sscli && cd sscli
 $ export SSCLI_PATH=$pwd
 $ cd cmd/sscli && go install
 $ sscli
 [GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
  - using env:   export GIN_MODE=release
  - using code:  gin.SetMode(gin.ReleaseMode)
 
 [GIN-debug] GET    /static/*filepath         --> github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)
 [GIN-debug] HEAD   /static/*filepath         --> github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)
 {"L":"INFO","T":"2019-03-23T11:45:31.057+0800","C":"sscli/sscli.go:99","M":"[http] listening => localhost:8888, serve => paths:[/static], dirs:[./]"}
 ```
 ### multi static dirs example
 ##### first start serve
  ```
  $ sscli -p /css -d /opt/css -p /img -d /opt/img
  {"L":"INFO","T":"2019-03-23T11:51:28.073+0800","C":"sscli/sscli.go:33","M":"dir /opt/css not found, create now..."}
  {"L":"INFO","T":"2019-03-23T11:51:28.074+0800","C":"sscli/sscli.go:33","M":"dir /opt/img not found, create now..."}
  [GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
   - using env:   export GIN_MODE=release
   - using code:  gin.SetMode(gin.ReleaseMode)
  
  [GIN-debug] GET    /css/*filepath            --> github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)
  [GIN-debug] HEAD   /css/*filepath            --> github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)
  [GIN-debug] GET    /img/*filepath            --> github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)
  [GIN-debug] HEAD   /img/*filepath            --> github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)
  {"L":"INFO","T":"2019-03-23T11:51:28.074+0800","C":"sscli/sscli.go:98","M":"[http] listening => localhost:8888, serve => paths:[/css /img], dirs:[/opt/css /opt/img]"}
  ```
  ##### then open another shell to test
   ```
   $ echo "test111111111111112222333" >> /opt/css/test
   $ curl localhost:8888/css/test
   test111111111111112222333
   ``` 
