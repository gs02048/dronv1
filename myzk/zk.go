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
)
var (
	// error
	ErrNoChild      = errors.New("zk: children is nil")
	ErrNodeNotExist = errors.New("zk: node not exist")
)

type Register struct {
	Conn *zk.Conn
	Prefix string
	WatchEvent chan bool
}

func NewRegister(conn *zk.Conn)(*Register,error){
	return &Register{Conn:conn,WatchEvent:make(chan bool,1)},nil
}

func (r *Register) RegisterTask(prefix string,task *task.Task)(string,error){
	if err := createPath(prefix,[]byte(""),r.Conn);err != nil{
		return "",err
	}
	var cpath string
	path := prefix+"/"+task.TaskName
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

func (r *Register)GetTask(prefix string,name string)(*task.Task,error){
	info,_,err := r.Conn.Get(prefix+"/"+name)
	if err != nil{
		return nil,err
	}
	item := &task.Task{}
	if err = json.Unmarshal(info,item);err != nil{
		return nil,err
	}
	if len(item.TaskName) <= 0{
		return nil,ErrNodeNotExist
	}
	return item,nil
}

func (r *Register)ListTask(prefix string)([]*task.Task,error){
	chs,_,err := r.Conn.Children(prefix)
	if err != nil{
		return nil,err
	}
	tasklist := make([]*task.Task,len(chs))

	for k,v := range(chs){
		item := &task.Task{}
		line,_,err := r.Conn.Get(prefix+"/"+v)
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

func (r *Register) ListService(prefix string)([]string,error){
	chs,_,err := r.Conn.Children(prefix)
	if err != nil{
		return nil,err
	}
	if len(chs) <= 0 {
		return nil,ErrNoChild
	}
	servicelist := make([]string,len(chs))
	for k,v := range(chs){
		servicelist[k] = v
	}
	return servicelist,nil
}

func (r *Register) RegisterService(prefix,name string,data []byte)(string,error){
	if err := createPath(prefix,[]byte(""),r.Conn);err != nil{
		return "",err
	}
	path := prefix+"/"+name
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

func (r *Register) GetDirW(path string){
	for{
		_,_,ch,err := r.Conn.ChildrenW(path)
		if err != nil{
			return
		}
		select{
			case  <-ch:
				//if e.Type == zk.EventNodeChildrenChanged{
					r.WatchEvent <- true
				//}
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