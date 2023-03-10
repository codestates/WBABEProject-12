package main

import (
	"context"
	"flag"
	"fmt"
	conf "go-ordering/conf"
	ctl "go-ordering/controller"
	"go-ordering/logger"
	"go-ordering/model"
	rt "go-ordering/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

/* 아래 항목이 swagger에 의해 문서화 된다. */
// @title WBA [Backend Final Project]
// @version 1.0
// @description 띵동주문이요, 온라인 주문 시스템(Online Ordering System)
func main() {

	//conf
	var configFlag = flag.String("config", "./conf/config.toml", "toml file to use for configuration")
	flag.Parse()
	cf := conf.GetConfig(*configFlag)

	// 로그 초기화
	if err := logger.InitLogger(cf); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	logger.Debug("ready server....")

	//model 모듈 선언
	fmt.Println("main.start")
	if mod, err := model.NewModel(); err != nil {
		//~생략
		fmt.Println("main.model.NewModel : ", err)
		panic(err)
	} else if controller, err := ctl.NewCTL(mod); err != nil { //controller 모듈 설정
		//~생략
		fmt.Println("main.ctl.NewCTL : ", err)
		panic(err)
	} else if rt, err := rt.NewRouter(controller); err != nil { //router 모듈 설정
		//~생략
		fmt.Println("main.rt.NewRouter : ", err)
		panic(err)
	} else {
		fmt.Println("main else ")
		mapi := &http.Server{
			//~생략
			Addr:           cf.Server.Port,
			Handler:        rt.Idx(),
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		go func() {
			if err := mapi.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()

		stopSig := make(chan os.Signal) //chan 선언
		// 해당 chan 핸들링 선언, SIGINT, SIGTERM에 대한 메세지 notify
		signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)
		<-stopSig //메세지 등록
		logger.Warn("Shutdown Server ...")

		// 해당 context 타임아웃 설정, 5초후 server stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mapi.Shutdown(ctx); err != nil {
			//log.Fatal("Server Shutdown:", err)
			logger.Error("Server Shutdown:", err)
		}
		// catching ctx.Done(). timeout of 5 seconds.
		select {
		case <-ctx.Done():
			//fmt.Println("timeout 5 secondse.")
			logger.Info("timeout of 5 seconds.")
		}
		//fmt.Println("Server stop")
		logger.Info("Server exiting")

	}

	if err := g.Wait(); err != nil {
		//fmt.Println("main.g.Wait : ", err)
		logger.Error(err)
	}

	fmt.Println("Server Stop")

}
