package worker

import (
	"net/http"
	"time"
	"net"
	"fmt"
	"dronv1/drond"
)

func Main(cfg *drond.Config){
	mux := http.NewServeMux()
	worker := NewWorker(cfg)
	mux.HandleFunc("/task",worker.RunTask)

	server := http.Server{Handler:mux,ReadTimeout:1*time.Second,WriteTimeout:1*time.Second}
	addr,_ := extractAddress()
	listener,err := net.Listen("tcp",addr+":")
	if err != nil{
		fmt.Println(err)
		return
	}
	//注册节点
	worker.Register.RegisterService(worker.Config.ZkServicePrefix,listener.Addr().String(),[]byte(""))

	server.Serve(listener)
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

func (t *Worker)RunTask(w http.ResponseWriter,r *http.Request){
	name := r.URL.Query().Get("name")
	if len(name) <= 0 {
		return
	}
	task,err := t.Register.GetTask(t.Config.ZkTaskPrefix,name)
	if err != nil{
		return
	}
	t.Run(task)
}
