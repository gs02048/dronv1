package drond

import (
	"dronv1/myzk"
	"time"
	log "github.com/alecthomas/log4go"
	"dronv1/cron"
	"math/rand"
	"net/http"
	"io/ioutil"
)

type DROND struct {
	Elect *myzk.Election
	Register *myzk.Register
	ServiceList []string
	Cfg *Config
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

	return &DROND{Register:register,Cfg:cfg,Elect:elect,ServiceList:servicelist}
}

func (d *DROND)watchService(){
	for{
		select {
			case e := <-d.Register.WatchEvent:
				if e{
					d.ServiceList,_ = d.Register.ListService(d.Cfg.ZkServicePrefix)
				}
		}
	}
}

func (d *DROND)Main(){
	rand.Seed(time.Now().UnixNano())
	go d.Elect.ElectMaster(d.Cfg.Electpre,d.Cfg.Electpath)
	go d.watchService()
	for{
		select {
			case m := <-d.Elect.IsMaster:
				c := cron.New()
				if m{
					log.Info("i am master")
					tasklist,_ := d.Register.ListTask(d.Cfg.ZkTaskPrefix)
					for _,task := range(tasklist){
						c.AddFunc(task.Spec,task.TaskName,func(){
							d.runRemote(task.TaskName)
						})
					}
					c.Run()

				}else{
					log.Info("i am slave")
				}
		}
	}
}

func (d *DROND)runRemote(name string){
	if len(d.ServiceList) <= 0 {
		log.Info("no service")
		return
	}
	i := rand.Int() % len(d.ServiceList)
	addr := d.ServiceList[i]
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