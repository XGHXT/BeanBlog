package main

import (
	"BeanBlog/server"
	"fmt"
)

func main() {
	endRun := make(chan error, 1)
	server.Serve(endRun)
	fmt.Println(<-endRun)
}
