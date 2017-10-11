package myzk

import "github.com/samuel/go-zookeeper/zk"

type Election struct {
	Conn *zk.Conn
	IsMaster chan bool
}

func NewElection(conn *zk.Conn) (*Election,error){
	return &Election{Conn:conn,IsMaster:make(chan bool,1)},nil
}

func (e *Election) ElectMaster(prefix,path string)(err error){
	if err := createPath(prefix,[]byte(""),e.Conn);err != nil{
		return err
	}
	electpath := prefix+"/"+path
	var cpath string
elect:
	cpath,err = e.Conn.Create(electpath,nil,zk.FlagEphemeral,zk.WorldACL(zk.PermAll))
	if err != nil || cpath != electpath{
		e.IsMaster <- false
	}else{
		e.IsMaster <- true
	}

	for{
		_,_,ch,err := e.Conn.ChildrenW(electpath)
		if err != nil{
			return err
		}
		select {
		case childEvent := <-ch:
			if childEvent.Type == zk.EventNodeDeleted{
				goto elect
			}
		}
	}
	return nil
}




