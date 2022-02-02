package core

import . "getitle/src/utils"

var InterConfig = map[string][]string{
	"10.0.0.0/8":     {"ss", "icmp", "1"},
	"172.16.0.0/12":  {"ss", "icmp", "1"},
	"192.168.0.0/16": {"s", "80", "all"},
	"100.100.0.0/16": {"s", "icmp", "all"},
	"200.200.0.0/16": {"s", "icmp", "all"},
	//"169.254.0.0/16": {"s", "icmp", "all"},
	//"168.254.0.0/16": {"s", "icmp", "all"},
}

type Options struct {
	AliveSum    int
	Noscan      bool
	Compress    bool
	Debug       bool
	file        *File
	smartFile   *File
	extractFile *File
	aliveFile   *File
	dataCh      chan string
	extractCh   chan string
	Output      string
	FileOutput  string
	FilePath    string
}

var Log *Logger

func (opt *Options) Close() {
	// 关闭管道
	close(Opt.dataCh)
	close(Opt.extractCh)
}
