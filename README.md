


## Installation 



```bash
go mod init github.com/mascot

go get -u github.com/go-gl/gl/v4.1-core/gl
go get -u github.com/go-gl/glfw/v3.3/glfw

```

_(Optional)_ If you really want backwards / compatibility profiles or GLES for testing on GLES devices, you can pull in those too:

```bash
go get -u github.com/go-gl/gl/v4.1-compatibility/gl
go get -u github.com/go-gl/gl/v3.3/glfw         # already above
go get -u github.com/go-gl/gl/v3.1/gles2
go get -u github.com/go-gl/gl/v2.1/gl

```




## Run 

update `main.go` with working sprites 


```bash
go run .

```
