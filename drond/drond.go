package drond

import (
	"dronv1/myzk"
	"time"
	log "github.com/alecthomas/log4go"
	"dronv1/cron"
	"math/rand"
	"net/http"
	"io/ioutil"
	"fmt"
)
var ServiceList []string

type DROND struct {
	Elect *myzk.Election
	Register *myzk.Register
	ServiceList []string
	Cfg *Config
	TaskEvent chan bool
}

type Config struct {
	ZkAddrs []string
	ZkTimeout time.Duration
	ZkServicePrefix string
	ZkTaskPrefix string

	Electpre string
	Electpath string
}

func New(cfg *Config)*DROND{
	zkconn,err := myzk.Connect(cfg.ZkAddrs,cfg.ZkTimeout)
	if err != nil{
		return nil
	}
	register,err := myzk.NewRegister(zkconn)
	elect,err := myzk.NewElection(zkconn)
	servicelist,err := register.ListService(cfg.ZkServicePrefix)
	go register.GetDirW(cfg.ZkServicePrefix)
	if err != nil{
		servicelist = []string{}
	}

	return &DROND{Register:register,Cfg:cfg,Elect:elect,ServiceList:servicelist,TaskEvent:make(chan bool,1)}
}

func (d *DROND)watchService(){
	for{
		select {
			case e := <-d.Register.WatchEvent:
				if e{
					ServiceList,_ = d.Register.ListService(d.Cfg.ZkServicePrefix)
				}
		}
	}
}

func (d *DROND)Main(){
	ServiceList,_ = d.Register.ListService(d.Cfg.ZkServicePrefix)
	rand.Seed(time.Now().UnixNano())
	go d.Elect.ElectMaster(d.Cfg.Electpre,d.Cfg.Electpath)
	go d.watchService()
	go d.Register.GetDirWatchEvent(d.Cfg.ZkTaskPrefix,d.TaskEvent)
	for{
		select {
			case m := <-d.Elect.IsMaster:
				cronstart:
				c := cron.New()
				if m{
					log.Info("i am master")
					tasklist,_ := d.Register.ListTask(d.Cfg.ZkTaskPrefix)
					for _,task := range(tasklist){
						c.AddFunc(task.Spec,task.TaskName,func(){})
					}

					c.Run()

					te := <-d.TaskEvent
					if te{
						c.Stop()
						fmt.Println("cron restart")
						goto cronstart
					}

				}else{
					log.Info("i am slave")
					c.Stop()
				}
		}
	}
}

func (d *DROND)watchCronDir(){
	go d.Register.GetDirWatchEvent(d.Cfg.ZkTaskPrefix,d.TaskEvent)
	for{
		select {
		case e:=<-d.TaskEvent:
			if e{

			}

		}
	}
}

func RunRemote(name string){
	if len(ServiceList) <= 0 {
		log.Info("no service")
		return
	}
	i := rand.Int() % len(ServiceList)
	addr := ServiceList[i]
	req,err := http.Get("http://"+addr+"/task?name="+name)
	if err != nil{
		log.Info(err)
		return
	}
	defer req.Body.Close()
	body,err := ioutil.ReadAll(req.Body)
	if err != nil{
		log.Info(err)
		return
	}
	log.Info("resp:%s",string(body))

}