package admin

import (
	"net/http"
	"time"
	"net"
	log"github.com/alecthomas/log4go"

	"encoding/json"
	"fmt"
	"dronv1/myzk"
	"dronv1/task"
	"strconv"
)
type Admin struct {
	Cfg *Config
	Register *myzk.Register
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
	register,err := myzk.NewRegister(zkconn)
	return &Admin{Cfg:cfg,Register:register}
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
	list,err := a.Register.ListTask(a.Cfg.ZkServicePrefix)
	if err != nil{

	}
	ret := map[string]interface{}{"ret":"ok"}
	ret["list"] = list
	defer retWrite(w,r,ret,"",time.Now())
	return
}

func (a *Admin)CreateService(w http.ResponseWriter,r *http.Request){
	if r.Method != "GET"{
		http.Error(w,"method not allowed",405)
		return
	}
	params := r.URL.Query()
	taskname := params.Get("taskname")
	desc := params.Get("desc")
	cmd := params.Get("command")
	args := params.Get("args")
	path := params.Get("path")
	spec := params.Get("spec")
	ty := params.Get("tasktype")
	var tasktype,maxruntime int64
	tasktype,_ = strconv.ParseInt(ty,10,64)
	mrt := params.Get("maxruntime")
	maxruntime,_ = strconv.ParseInt(mrt,10,64)

	res := map[string]interface{}{"ret":"ok"}
	defer retWrite(w,r,res,"",time.Now())
	if taskname == "" || cmd == "" || desc == "" || spec == "" || maxruntime <= 0{
		res["ret"] = "param err"
		return
	}

	t := &task.Task{
		TaskName:taskname,
		Desc:desc,
		Command:cmd,
		Args:args,
		Path:path,
		Spec:spec,
		TaskType:tasktype,
		MaxRunTime:maxruntime,
	}
	cpath,err := a.Register.RegisterTask(a.Cfg.ZkTaskPrefix,t)
	res["cpath"] = cpath
	res["err"] = err
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


