package tail

import (
	log "github.com/cihub/seelog"
	"io"
	"os"
	"runtime"
)

import (
	"github.com/axgle/mahonia"
)

type ReadFile struct {
	file    *os.File
	gbkFile *mahonia.Reader
}

func (f *ReadFile) gbkDecode() {
	decoder := mahonia.NewDecoder("gbk")
	f.gbkFile = decoder.NewReader(f.file)
}

func (f *ReadFile) ReadPrint() {
	var n int
	var err error
	data := make([]byte, 1<<16)
	if runtime.GOOS == "windows" {
		f.gbkDecode()
		n, err = f.gbkFile.Read(data)
	} else {
		n, err = f.file.Read(data)
	}
	switch err {
	case nil:
		var lines int
		out := data
		indexs := make(map[int]int)
		for i, d := range out {
			if d == '\n' {
				lines++
				indexs[lines] = i
			}
		}
		lines += 1

		if lines <= line || line <= 0 {
			logs := Json2Struct(string(data[:n]))
			mainLog := MainLogSync(logs)
			log.Info(mainLog.time)
			//mianLogStr := logs.Log

			//fmt.Println(util.UTCTransLocal(logs.Time))
		} else {
			//index := indexs[lines-line]
			//log := Json2Struct(string(data[index+1 : n]))
			//log.Info("xx", log)
		}

	case io.EOF:
	default:
		log.Info(err)
		return
	}
}

var (
	follow bool
	line   int
)

func init() {
	follow = true
	line = 10
}

func Stream(path string, m chan string) {
	var err error
	var readFile ReadFile
	readFile.file, err = os.Open(path)
	//readFile.file, err = os.Open(flag.Arg(0))
	if err != nil {
		log.Info(err)
		return
	}

	defer readFile.file.Close()
	defer close(m)
	for {
		readFile.ReadPrint()
		if !follow {
			m <- "exit"
			break
		}
	}
}
