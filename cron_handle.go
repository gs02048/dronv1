package main

import (
	"net/http"
	"time"
	"net"
	log"github.com/alecthomas/log4go"
	"io"
	"fmt"
	"dron/zk"
	"path"
)

func InitHttp(){
	httpServeMux := http.NewServeMux()
	httpServeMux.HandleFunc("/v1/handle",HandleCommand)

	httpServer := &http.Server{Handler:httpServeMux,ReadTimeout:1*time.Second,WriteTimeout:1*time.Second}
	httpServer.SetKeepAlivesEnabled(true)
	ip,_ := extractAddress()
	l,_ := net.Listen("tcp",ip+":")
	log.Info("http address:%s",l.Addr().String())
	zk.RegisterTemp(ZkConn,path.Join(zk.CronService,l.Addr().String()),[]byte(""))
	if err:=httpServer.Serve(l);err != nil{
		log.Error("http serve err:%s",err)
		panic(err)
	}
}


func HandleCommand(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"hello")
}


func extractAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("Failed to get interface addresses! Err: %v", err)
	}

	for _, rawAddr := range addrs {
		if ipnet, ok := rawAddr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return  ipnet.IP.String(),nil
			}
		}
	}
	return "",nil
}

