package admin

import (
	"net/http"
	"time"
	"net"
	log"github.com/alecthomas/log4go"

	"encoding/json"
	"io/ioutil"
	"fmt"
	"errors"
	"dronv1/myzk"
	"dronv1/task"
)
type Admin struct {
	Cfg *Config
	ServiceReg *myzk.Register
	TaskReg *myzk.Register
}

type Config struct {
	HttpPort string
	ZkAddrs []string
	ZkTimeout time.Duration
	ZkServicePrefix string
	ZkTaskPrefix string
}


func NewAdmin(cfg *Config)*Admin{
	zkconn,err := myzk.Connect(cfg.ZkAddrs,cfg.ZkTimeout)
	if err != nil{
		log.Error(err)
		return nil
	}
	service_reg,err := myzk.NewRegister(zkconn,cfg.ZkServicePrefix)
	task_reg,err := myzk.NewRegister(zkconn,cfg.ZkTaskPrefix)
	return &Admin{Cfg:cfg,ServiceReg:service_reg,TaskReg:task_reg}
}

func InitAdminHttp(a *Admin){
	httpServeMux := http.NewServeMux()
	httpServeMux.HandleFunc("/v1/list",a.ListService)
	httpServeMux.HandleFunc("/v1/put",a.CreateService)

	httpServer := &http.Server{Handler:httpServeMux,ReadTimeout:1*time.Second,WriteTimeout:1*time.Second}
	httpServer.SetKeepAlivesEnabled(true)
	log.Info(":"+a.Cfg.HttpPort)
	l,_ := net.Listen("tcp","localhost:"+a.Cfg.HttpPort)
	log.Info("http address:%s",l.Addr().String())
	if err:=httpServer.Serve(l);err != nil{
		log.Error("http serve err:%s",err)
		panic(err)
	}
}



func (a *Admin)ListService(w http.ResponseWriter,r *http.Request){
	list,err := a.TaskReg.ListTask()
	if err != nil{

	}
	ret := map[string]interface{}{"ret":"ok"}
	ret["list"] = list
	defer retWrite(w,r,ret,"",time.Now())
	return
}

func (a *Admin)CreateService(w http.ResponseWriter,r *http.Request){
	ret := map[string]interface{}{"ret":"ok"}
	b,err := ioutil.ReadAll(r.Body)
	if err != nil{
		ret["err"] = err
		retWrite(w,r,ret,"",time.Now())
		return
	}
	c := &task.Task{}
	if err := json.Unmarshal(b,c);err != nil{
		ret["err"] = err
		retWrite(w,r,ret,"",time.Now())
		return
	}
	log.Debug(c)
	if c.TaskName == ""{
		ret["err"] = errors.New("cron name error")
		retWrite(w,r,ret,"",time.Now())
		return
	}
	a.TaskReg.RegisterTask(c)
	retWrite(w,r,ret,"",time.Now())
	return
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


