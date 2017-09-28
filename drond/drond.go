package drond

import (
	"dronv1/myzk"
	"time"
	log "github.com/alecthomas/log4go"
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
	go d.Elect.ElectMaster(d.Cfg.Electpre,d.Cfg.Electpath)
	for{
		select {
			case m := <-d.Elect.IsMaster:

			if m{
				log.Info("i am master")
			}else{
				log.Info("i am slave")
			}
		}
	}
}