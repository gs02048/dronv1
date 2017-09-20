package main

import (
	"net/http"
	"time"
	"net"
	log"github.com/alecthomas/log4go"

	"dronv1/zk"
	"fmt"
	"encoding/json"
)

func InitAdminHttp(port string){
	httpServeMux := http.NewServeMux()
	httpServeMux.HandleFunc("/v1/list",ListService)
	httpServeMux.HandleFunc("/v1/put",PutService)

	httpServer := &http.Server{Handler:httpServeMux,ReadTimeout:1*time.Second,WriteTimeout:1*time.Second}
	httpServer.SetKeepAlivesEnabled(true)
	ip,_ := extractAddress()
	l,_ := net.Listen("tcp",ip+":"+port)
	log.Info("http address:%s",l.Addr().String())
	if err:=httpServer.Serve(l);err != nil{
		log.Error("http serve err:%s",err)
		panic(err)
	}
}


func ListService(w http.ResponseWriter,r *http.Request){
	list,err := zk.ListCron(ZkConn)
	if err != nil{

	}
	ret := map[string]interface{}{"ret":"ok"}
	ret["list"] = list
	defer retWrite(w,r,ret,"",time.Now())
	return
}

func PutService(w http.ResponseWriter,r *http.Request){
	
}

func retWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, callback string, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := ""
	if callback == "" {
		// Normal json
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		dataStr = string(data)
	} else {
		// Jsonp
		w.Header().Set("Content-Type", "application/javascript;charset=utf-8")
		dataStr = fmt.Sprintf("%s(%s)", callback, string(data))
	}
	if n, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	} else {
		log.Debug("w.Write(\"%s\") write %d bytes", dataStr, n)
	}
	log.Info("req: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}


