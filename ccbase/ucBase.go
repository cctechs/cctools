package ccbase

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cctechs/cctools/cclog"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
)

// 打印系统相关的数据
func PrintSysInfo(){
	fmt.Printf("cpu num:%d\n", runtime.NumCPU())
	fmt.Printf("go version:%v\n", runtime.Version())
}

func Wait()  {
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGTERM | syscall.SIGKILL | syscall.SIGQUIT)
	for{
		select {
		case s := <- chSignal:
			cclog.LogNotice("Receive os.signal:%v, Exited", s)
			time.Sleep(time.Second)  // wait for writing log
			os.Exit(0)
		}
	}
}

// 获取数据的字符流长度
func StreamSizeof(data interface{}) uint32{
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, data)
	return uint32(buf.Len())
}

// 延迟执行
func RunAfter(duration time.Duration, handler interface{}){
	go func() {
		timer := time.NewTimer(duration)
		for{
			select {
			case <- timer.C:
				{
					cb, ok := handler.(func())
					if ok{
						cb()
					}
					return
				}
			}
		}
	}()
}

// 序列化
func SerializeToBytes(data interface{}) []byte{
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, data)
	if err != nil{
		panic(err.Error())
	}
	return buf.Bytes()
}

// 反序列化
func UnSerializeFromBytes(buf []byte, data interface{}) error{
	buffer := new(bytes.Buffer)
	buffer.Write(buf)
	err := binary.Read(buffer, binary.BigEndian, data)
	if err != nil{
		panic(err.Error())
	}
	return nil
}


func PProf(file string){
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
}