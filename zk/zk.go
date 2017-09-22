package zk

import (
	log "github.com/alecthomas/log4go"
	"errors"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
	"encoding/json"
	"dronv1/define"
)

var (
	// error
	ErrNoChild      = errors.New("zk: children is nil")
	ErrNodeNotExist = errors.New("zk: node not exist")
	prefix = "/LCSCRON"
	CronService = "/CRONSERVICE"
	MasterCron = "/ElectMasterCron/master"
)

// Connect connect to zookeeper, and start a goroutine log the event.
func Connect(addr []string, timeout time.Duration,cp bool) (*zk.Conn, error) {
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
	if cp {
		if err := createPath(prefix,[]byte(""),conn);err != nil{
			log.Error("connect create path err:%s",err)
			return nil,err
		}
		if err := createPath(CronService,[]byte(""),conn);err != nil{
			log.Error("connect create path err:%s",err)
			return nil,err
		}
	}
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


func RegisterCron(conn *zk.Conn,fpath string,data []byte)(error){
	cron := &define.CronLine{}
	if err := json.Unmarshal(data,cron);err != nil{
		log.Error("json.Unmarshal cronline err:%s",err)
		return err
	}
	tpath,err := conn.Create(prefix+"/"+cron.Name,data,int32(0), zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Error("conn.Create(\"%s\", \"%s\", zk.FlagEphemeral|zk.FlagSequence) error(%v)", fpath, string(data), err)
		return err
	}
	log.Debug("create a zookeeper node:%s", tpath)
	return nil

}

func ListCron(conn *zk.Conn)([]*define.CronLine,error){
	chs,_,err := conn.Children(prefix)
	if err != nil{
		log.Error("children cron err:%s",err)
		return nil,err
	}
	cronlist := make([]*define.CronLine,len(chs))
	item := &define.CronLine{}
	for k,v := range(chs){

		line,_,err := conn.Get(prefix+"/"+v)
		if err != nil{
			log.Error("list cron err:%s",err)
			continue
		}
		if err = json.Unmarshal(line,item);err != nil{
			log.Error("list cron json unmarshal err:%s",err)
			continue
		}
		cronlist[k] = item
	}
	return cronlist,nil
}

// RegisterTmp create a ephemeral node, and watch it, if node droped then send a SIGQUIT to self.
func RegisterTemp(conn *zk.Conn, fpath string, data []byte) error {
	tpath, err := conn.Create(fpath, data, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Error("conn.Create(\"%s\", \"%s\", zk.FlagEphemeral|zk.FlagSequence) error(%v)", fpath, string(data), err)
		return err
	}
	log.Debug("create a zookeeper node:%s", tpath)
	// watch self
	go func() {
		for {
			log.Info("zk path: \"%s\" set a watch", tpath)
			exist, _, watch, err := conn.ExistsW(tpath)
			if err != nil {
				log.Error("zk.ExistsW(\"%s\") error(%v)", tpath, err)
				log.Warn("zk path: \"%s\" set watch failed, kill itself", tpath)
				killSelf()
				return
			}
			if !exist {
				log.Warn("zk path: \"%s\" not exist, kill itself", tpath)
				killSelf()
				return
			}
			event := <-watch
			log.Info("zk path: \"%s\" receive a event %v", tpath, event)
		}
	}()
	return nil
}

// GetNodesW get all child from zk path with a watch.
func GetNodesW(conn *zk.Conn, path string) ([]string, <-chan zk.Event, error) {
	nodes, stat, watch, err := conn.ChildrenW(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, nil, ErrNodeNotExist
		}
		log.Error("zk.ChildrenW(\"%s\") error(%v)", path, err)
		return nil, nil, err
	}
	if stat == nil {
		return nil, nil, ErrNodeNotExist
	}
	if len(nodes) == 0 {
		return nil, nil, ErrNoChild
	}
	return nodes, watch, nil
}

// GetNodes get all child from zk path.
func GetNodes(conn *zk.Conn, path string) ([]string, error) {
	nodes, stat, err := conn.Children(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, ErrNodeNotExist
		}
		log.Error("zk.Children(\"%s\") error(%v)", path, err)
		return nil, err
	}
	if stat == nil {
		return nil, ErrNodeNotExist
	}
	if len(nodes) == 0 {
		return nil, ErrNoChild
	}
	return nodes, nil
}

func ElectMaster(conn *zk.Conn,ismaster chan bool)(error){
	isExist,_,err := conn.Exists("/ElectMasterCron")
	if err != nil{
		return err
	}
	if !isExist{
		err = createPath("/ElectMasterCron",[]byte(""),conn)
		if err != nil{
			return err
		}
	}
	path,err := conn.Create(MasterCron,nil,zk.FlagEphemeral,zk.WorldACL(zk.PermAll))
	if err == nil{
		if path == MasterCron{
			log.Info("elect master success!")
			ismaster <- true
		}else{
			return errors.New("return path diff")
		}
	}else{
		log.Info("elect master failure err:%s",err)
		ismaster <- false
	}

	return nil
}

func WatchMaster(conn *zk.Conn,ismaster chan bool){
	for{
		children,state,childCh,err := conn.ChildrenW(MasterCron)
		if err != nil{
			log.Error("watch children error:%s",err)
		}
		log.Info("watch children result:%s,state:%s",children,state)
		select {
			case childEvent := <-childCh:
				if childEvent.Type == zk.EventNodeDeleted{
					log.Info("receive znode delete event:%s",childEvent)
					log.Info("start elect new master ...")
					err = ElectMaster(conn,ismaster)
					if err != nil{
						log.Error("elect new master error:%s",err)
					}
				}
		}
	}
}

// killSelf send a SIGQUIT to self.
func killSelf() {
	if err := syscall.Kill(os.Getpid(), syscall.SIGQUIT); err != nil {
		log.Error("syscall.Kill(%d, SIGQUIT) error(%v)", os.Getpid(), err)
	}
}
