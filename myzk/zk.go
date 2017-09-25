package myzk

import (
	log "github.com/alecthomas/log4go"
	"time"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"dronv1/task"
	"encoding/json"
	"syscall"
	"os"
	"errors"
	"fmt"
)
var (
	// error
	ErrNoChild      = errors.New("zk: children is nil")
	ErrNodeNotExist = errors.New("zk: node not exist")
)

type Register struct {
	Conn *zk.Conn
	Prefix string
}

func NewRegister(conn *zk.Conn,prefix string)(*Register,error){
	if err := createPath(prefix,[]byte(""),conn);err != nil{
		return nil,err
	}
	return &Register{Conn:conn,Prefix:prefix},nil
}

func (r *Register) RegisterTask(task *task.Task)(string,error){
	var cpath string
	path := r.Prefix+"/"+task.TaskName
	data,err := json.Marshal(task)
	if err != nil{
		return cpath,err
	}

	cpath,err = r.Conn.Create(path,data,int32(0),zk.WorldACL(zk.PermAll))
	if err != nil{
		return cpath,err
	}
	return cpath,nil
}

func (r *Register)ListTask()([]*task.Task,error){
	chs,_,err := r.Conn.Children(r.Prefix)
	if err != nil{
		return nil,err
	}
	tasklist := make([]*task.Task,len(chs))

	for k,v := range(chs){
		item := &task.Task{}
		line,_,err := r.Conn.Get(r.Prefix+"/"+v)
		if err != nil{
			continue
		}
		if err = json.Unmarshal(line,item);err != nil{
			continue
		}
		if len(item.TaskName) <= 0{
			continue
		}
		tasklist[k] = item
	}
	return tasklist,nil
}

func (r *Register) RegisterService(name string,data []byte)(string,error){
	path := r.Prefix+"/"+name
	cpath,err := r.Conn.Create(path,data,zk.FlagEphemeral,zk.WorldACL(zk.PermAll))
	if err != nil{
		return cpath,err
	}
	go func(){
		for{
			exist,_,watch,err := r.Conn.ExistsW(cpath)
			if err != nil{
				log.Warn("zk path: \"%s\" set watch failed, kill itself", cpath)
				KillSelf()
				return
			}
			if !exist{
				log.Warn("zk path: \"%s\" not exist, kill itself", cpath)
				KillSelf()
				return
			}
			event := <-watch
			log.Info("zk path: \"%s\" receive a event %v", cpath, event)
		}
	}()
	return cpath,nil
}

func (r *Register) GetNodesW(path string){
	for{
		nodes,_,ch,err := r.Conn.ChildrenW(path)
		if err != nil{
			if err == zk.ErrNoNode{

			}
		}
	}


}

func KillSelf(){
	if err := syscall.Kill(os.Getpid(),syscall.SIGQUIT);err != nil{
		log.Error("syscall.Kill(%d, SIGQUIT) error(%v)", os.Getpid(), err)
	}
}


// Connect connect to zookeeper, and start a goroutine log the event.
func Connect(addr []string, timeout time.Duration) (*zk.Conn, error) {
	conn, session, err := zk.Connect(addr, timeout)
	if err != nil {
		log.Error("zk.Connect(\"%v\", %d) error(%v)", addr, timeout, err)
		return nil, err
	}
	go func() {
		for {
			event := <-session
			log.Debug("zookeeper get a event: %s", event.State.String())
		}
	}()
	return conn, nil
}


func createPath(path string, data []byte, client *zk.Conn) error {
	exists, _, err := client.Exists(path)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	name := "/"
	p := strings.Split(path, "/")

	for _, v := range p[1 : len(p)-1] {
		name += v
		e, _, _ := client.Exists(name)
		if !e {
			_, err = client.Create(name, []byte{}, int32(0), zk.WorldACL(zk.PermAll))
			if err != nil {
				return err
			}
		}
		name += "/"
	}

	_, err = client.Create(path, data, int32(0), zk.WorldACL(zk.PermAll))
	return err
}