package gpool

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

//Pool pool class
type Pool struct {
	Config       *Config
	Items        *list.List
	cond         *sync.Cond
	context      context.Context
	shutdownFunc context.CancelFunc
	shutdownChan chan int
	NewFunc      func() Item
}

//Item pool item
type Item interface {
	Initial(map[string]string) error
	Destory() error
	Check() error
}

//DefaultPool create a pool with default config
func DefaultPool() *Pool {
	var result = &Pool{
		Config:       DefaultConfig(),
		cond:         sync.NewCond(&sync.Mutex{}),
		shutdownChan: make(chan int, 1),
	}
	result.context, result.shutdownFunc = context.WithCancel(context.Background())
	return result
}

//Initial initial pool
func (pool *Pool) Initial() {
	if pool.Config == nil {
		log.Fatal("pool config is nil")
	}

	pool.Log("START", "Pool Initial")

	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	pool.Items = list.New()

	//push item into pool
	go pool.Extend(pool.Config.InitialPoolSize)
	pool.cond.Wait()

	//start check avaiable goroutine
	go pool.StartCheck()

	pool.Log("DONE", "Pool Initial")
}

//Extend push item into pool
func (pool *Pool) Extend(count int) {
	pool.Log("START", fmt.Sprintf("Extend Count : %d", count))

	defer pool.cond.Signal()
	defer pool.cond.L.Unlock()
	pool.cond.L.Lock()
	if pool.Items.Len() >= pool.Config.MaxPoolSize {
		return
	}
	for i := 0; i < count; i++ {
		var item = pool.NewFunc()
		err := item.Initial(pool.Config.Params)
		if err != nil {
			log.Println("ERROR : Iem Initial ERROR \n", err)
			continue
		}
		pool.Items.PushBack(item)
	}

	pool.Log("DONE", fmt.Sprintf("Extend Count : %d ,Pool size : %d", count, pool.Items.Len()))
}

//StartCheck start check avaiable goroutine
func (pool *Pool) StartCheck() {
	t := time.NewTicker(time.Duration(pool.Config.TestDuration) * time.Millisecond)
a:
	for {
		select {
		case <-pool.context.Done():
			break a
		case <-t.C:
			pool.Log("START", "CheckAvaiable")
			pool.CheckAvaiable()
			pool.Log("DONE", "CheckAvaiable")
		}
	}
	pool.shutdownChan <- 1
}

//CheckAvaiable check item avaiable
func (pool *Pool) CheckAvaiable() {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	for i := pool.Items.Front(); i != nil; i = i.Next() {
		item, Ok := i.Value.(Item)
		if !Ok {
			log.Println("ERROR : Class Cast ERROR while CheckAvaiable")
		}
		err := item.Check()
		if err != nil {
			log.Println("ERROR : CheckAvaiable ERROR \n", err)
			pool.Items.Remove(i)
		}
	}
}

//GetOne get a pool item
func (pool *Pool) GetOne() (Item, error) {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	retry := 0
	for {
		//检查链表头元素是否为空，防止链表遍历结束依然未获取到连接时报错
		i := pool.Items.Front()
		if i == nil {
			if retry <= pool.Config.AcquireRetryAttempts {
				retry++
				go pool.Extend(pool.Config.AcquireIncrement)
				pool.cond.Wait()
				continue
			}
			return nil, errors.New("Unable GET Item")
		}
		item, ok := i.Value.(Item)
		pool.Items.Remove(i)
		if !ok {
			return nil, errors.New("Class Cast ERROR while Get Item")
		}
		if pool.Items.Len() < pool.Config.MinPoolSize {
			go pool.Extend(pool.Config.AcquireIncrement)
			pool.cond.Wait()
		}
		if !pool.Config.TestOnGetItem {
			return item, nil
		}
		if item.Check() == nil {
			return item, nil
		}
	}
}

//BackOne  give back a pool item
func (pool *Pool) BackOne(item Item) {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	if pool.Items.Len() >= pool.Config.MaxPoolSize {
		err := item.Destory()
		if err != nil {
			log.Println("ERROR : Item Destory ERROR While BackOne \n", err)
		}
		return
	}
	pool.Items.PushBack(item)
	return
}

//Shutdown shutdown pool
func (pool *Pool) Shutdown() {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	pool.Log("START", "Shutdown Pool")
	for i := pool.Items.Front(); i != nil; i = pool.Items.Front() {
		item, ok := i.Value.(Item)
		pool.Items.Remove(i)
		if !ok {
			log.Println("ERROR : Class Cast ERROR while shutdown pool")
			continue
		}
		err := item.Destory()
		if err != nil {
			log.Println("ERROR : Item Destory ERROR While Shutdown \n", err)
		}
	}
	pool.shutdownFunc()
	<-pool.shutdownChan
	pool.Log("DONE", "Shutdown Pool")
}

//Log record log
func (pool *Pool) Log(status, msg string) {
	if pool.Config.Debug {
		log.Printf("INFO : [ %5s] %s\n", status, msg)
	}
}
