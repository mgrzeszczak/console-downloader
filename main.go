package main

import (
	"sync"
	"gopkg.in/cheggaaa/pb.v1"
	"net/http"
	"os"
	"io"
	"strconv"
	"fmt"
	"encoding/json"
)

type Files struct {
	Files []FileInfo `json:"files"`
}

type FileInfo struct {
	Url string `json:"url"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func downloadFile(fi FileInfo, wg *sync.WaitGroup, pb *pb.ProgressBar) error {
	defer func(){
		wg.Done()
	}()
	resp, err := http.Get(fi.Url)
	if err!=nil {
		return err
	}
	r := pb.NewProxyReader(resp.Body)

	f,err := os.Create(fi.Path)
	if err!=nil {
		return err
	}
	_,err = io.Copy(f,r)
	return err
}

func createBarForFile(fi FileInfo) *pb.ProgressBar {
	l := 0
	resp,err := http.Head(fi.Url)
	if err==nil {
		l,err = strconv.Atoi(resp.Header.Get("Content-Length"))
		if err!=nil {
			l = 0
		}
	}
	return pb.New(l).SetUnits(pb.U_BYTES).Prefix(fi.Name)
}

func downloadAll(fis []FileInfo) {
	pbs := make([]*pb.ProgressBar,0)
	for _,fi := range(fis) {
		pbs = append(pbs,createBarForFile(fi))
	}
	pool, err := pb.StartPool(pbs...)
	if err!=nil {
		panic(err)
	}
	wg := new(sync.WaitGroup)
	wg.Add(len(pbs))

	for i,fi := range(fis){
		go downloadFile(fi,wg,pbs[i])
	}

	wg.Wait()
	pool.Stop()
}

func usage(){
	fmt.Printf("%s fname\n",os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) !=2 {
		usage()
	}
	f, err := os.Open(os.Args[1])
	if err!=nil {
		panic(err)
	}

	files := &Files{}
	err = json.NewDecoder(f).Decode(files)
	if err!=nil {
		panic(err)
	}

	downloadAll(files.Files)
}