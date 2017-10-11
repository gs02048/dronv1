package worker

import (
	"dronv1/task"
	"time"
	"github.com/pborman/uuid"
	"fmt"
	"os/exec"
	"bytes"
	"errors"
	"strings"
	"encoding/json"
	"dronv1/drond"
	"dronv1/myzk"
)

/**
* 1、接收运行task请求 2、异步运行task 3、记录task运行结果
 */
type Worker struct {
	Count int32
	Runing map[string]*Job

	Config *drond.Config
	Elect *myzk.Election
	Register *myzk.Register
}


type Job struct {
	*task.Task

	Uuid string
	StartRunTime int64
	EndRunTime int64
	Stdout string
	Stderr string
}


func NewWorker(cfg *drond.Config)*Worker{
	runing := map[string]*Job{}
	zkconn,err := myzk.Connect(cfg.ZkAddrs,cfg.ZkTimeout)
	if err != nil{
		return nil
	}
	register,err := myzk.NewRegister(zkconn)
	elect,err := myzk.NewElection(zkconn)
	return &Worker{
		Count:0,
		Runing:runing,
		Config:cfg,
		Elect:elect,
		Register:register,
	}
}

func (w *Worker) Run(task *task.Task){
	Uuid := uuid.New()
	runningjob := &Job{StartRunTime:time.Now().Unix(),Task:task,Uuid:Uuid}
	w.Runing[Uuid] = runningjob
	w.Count++
	go w.RunScriptTask(runningjob)
}

func (w *Worker) RunScriptTask(job *Job){
	var cmd *exec.Cmd
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var err error

	cmd = exec.Command("/bin/sh","-c",job.Command,job.Args)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Start()

	err,_ = w.CmdRunWithTimeout(
		cmd,time.Duration(job.MaxRunTime)*time.Second,
	)
	if err != nil{
		job.Stderr = fmt.Sprintf("stderr:%s",err)
	}
	if len(stderr.String()) != 0{
		job.Stderr = stderr.String()
	}else{
		job.Stdout = strings.TrimRight(stdout.String(), "\n")
	}
	
	job.EndRunTime = time.Now().Unix()
	data,_ := json.Marshal(job)
	fmt.Println(string(data))
	delete(w.Runing,job.Uuid)
}

func (w *Worker) CmdRunWithTimeout(cmd *exec.Cmd, timeout time.Duration) (error, bool) {
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	var err error
	select {
	case <-time.After(timeout):
		// timeout
		if err = cmd.Process.Kill(); err != nil {

		}

		go func() {
			<-done // allow goroutine to exit
		}()
		return errors.New("exec timeout"), true
	case err = <-done:
		return err, false
	}
}

func (w *Worker) Stop(){

}


