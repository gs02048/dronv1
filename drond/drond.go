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
	return &DROND{Register:register,Cfg:cfg,Elect:elect}
}

func (d *DROND)Main(){
	rand.Seed(time.Now().UnixNano())
	go d.Elect.ElectMaster(d.Cfg.Electpre,d.Cfg.Electpath)
	for{
		select {
			case m := <-d.Elect.IsMaster:
				c := cron.New()
				if m{
					log.Info("i am master")
					tasklist,_ := d.Register.ListTask(d.Cfg.ZkTaskPrefix)
					for _,task := range(tasklist){
						c.AddFunc(task.Spec,task.TaskName,func(){
							d.callRemote(task.TaskName)
						})
					}
					c.Run()

				}else{
					log.Info("i am slave")
				}
		}
	}
}

func (d *DROND)callRemote(name string){
	list,err := d.Register.ListService(d.Cfg.ZkServicePrefix)
	if err != nil{
		log.Info(err)
		return
	}
	i := rand.Int() % len(list)
	addr := list[i]
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