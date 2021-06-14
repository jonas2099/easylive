package main

import (
	"github.com/haroldleong/easylive/server"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()
	log.SetLevel(log.TraceLevel)
	log.Infof(`
   _____   _____   ______              _    
  (_____) / ___ \ |  ___ \    /\      | |   
     _   | |   | || |   | |  /  \      \ \  
    | |  | |   | || |   | | / /\ \      \ \ 
 ___| |  | |___| || |   | || |__| | _____) )
(____/    \_____/ |_|   |_||______|(______/ 
	`)
	server := server.New()
	server.StartServe()
}
