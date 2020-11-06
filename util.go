package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/cheggaaa/pb/v3"
)

type CloudflareIPData struct {
	ip            net.IPAddr
	pingCount     int
	pingReceived  int
	recvRate      float32
	downloadSpeed float32
	pingTime      float32
}

func (cf *CloudflareIPData) getRecvRate() float32 {
	if cf.recvRate == 0 {
		pingLost := cf.pingCount - cf.pingReceived
		cf.recvRate = float32(pingLost) / float32(cf.pingCount)
	}
	return cf.recvRate
}

func ExportCsv(filePath string, data []CloudflareIPData) {
	fp, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("创建文件["+filePath+"]句柄失败,%v", err)
		return
	}
	defer fp.Close()
	w := csv.NewWriter(fp) //创建一个新的写入文件流
	w.Write([]string{"IP 地址", "已发送", "已接收", "丢包率", "平均延迟", "下载速度 (MB/s)"})
	w.WriteAll(convertToString(data))
	w.Flush()
}

func (cf *CloudflareIPData) toString() []string {
	result := make([]string, 6)
	result[0] = cf.ip.String()
	result[1] = strconv.Itoa(cf.pingCount)
	result[2] = strconv.Itoa(cf.pingReceived)
	result[3] = strconv.FormatFloat(float64(cf.getRecvRate()), 'f', 2, 32)
	result[4] = strconv.FormatFloat(float64(cf.pingTime), 'f', 2, 32)
	result[5] = strconv.FormatFloat(float64(cf.downloadSpeed)/1024/1024, 'f', 2, 32)
	return result
}

func convertToString(data []CloudflareIPData) [][]string {
	result := make([][]string, 0)
	for _, v := range data {
		result = append(result, v.toString())
	}
	return result
}

var pingTime int
var pingRoutine int

var ipEndWith uint8 = 0

type progressEvent int

const (
	NoAvailableIPFound progressEvent = iota
	AvailableIPFound
	NormalPing
)

const url string = "https://uspf.myfalcom.website/test"

var downloadTestTime time.Duration

const downloadBufferSize = 1024

var downloadTestCount int

//const defaultTcpPort = 443
const tcpConnectTimeout = time.Second * 1

var failTime int

type CloudflareIPDataSet []CloudflareIPData

func initipEndWith() {
	rand.Seed(time.Now().UnixNano())
	ipEndWith = uint8(rand.Intn(254) + 1)
}

func handleProgressGenerator(pb *pb.ProgressBar) func(e progressEvent) {
	return func(e progressEvent) {
		switch e {
		case NoAvailableIPFound:
			pb.Add(pingTime)
		case AvailableIPFound:
			pb.Add(failTime)
		case NormalPing:
			pb.Increment()
		}
	}
}

func (cfs CloudflareIPDataSet) Len() int {
	return len(cfs)
}

func (cfs CloudflareIPDataSet) Less(i, j int) bool {
	if (cfs)[i].getRecvRate() != cfs[j].getRecvRate() {
		return cfs[i].getRecvRate() < cfs[j].getRecvRate()
	}
	return cfs[i].pingTime < cfs[j].pingTime
}

func (cfs CloudflareIPDataSet) Swap(i, j int) {
	cfs[i], cfs[j] = cfs[j], cfs[i]
}
