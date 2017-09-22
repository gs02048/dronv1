package myzk

import "github.com/samuel/go-zookeeper/zk"

type Election struct {
	Conn *zk.Conn
	Prefix string
	Path string
	IsMaster chan bool
}

func NewElection(conn *zk.Conn,prefix string) (*Election,error){
	if err := createPath(prefix,[]byte(""),conn);err != nil{
		return nil,err
	}
	return &Election{Conn:conn,Prefix:prefix},nil
}

func (e *Election) ElectMaster(path string)(err error){
	e.Path = path
	electpath := e.Prefix+"/"+path
	var cpath string
elect:
	cpath,err = e.Conn.Create(electpath,nil,zk.FlagEphemeral,zk.WorldACL(zk.PermAll))
	if err != nil || cpath != path{
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




