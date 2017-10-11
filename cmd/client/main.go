package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
)

func main(){
	resp,_ := http.Get("http://localhost:5890/v1/put?taskname=test&desc=test&command=ls ../&spec=* * * * * *&tasktype=1&maxruntime=10")
	defer resp.Body.Close()
	data,_ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(data))
}
