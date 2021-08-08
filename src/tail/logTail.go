package tail

import (
	log "github.com/cihub/seelog"
	"io"
	"os"
	"runtime"
	"time"
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

func (f *ReadFile) ReadPrint() int {
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
		var lines = 0
		out := data
		for _, d := range out {
			if d == '\n' {
				lines++
				lines += 1
			}
		}

		if lines <= line || line <= 0 {
			logStr := string(data[:n])
			//fmt.Println(logStr)
			sync := MainLogSync(logStr, time.Now().String())
			return sync
		}

	case io.EOF:
	default:
		log.Info(err)
		return 0
	}
	return 0
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
	if err != nil {
		log.Info(err)
		return
	}
	defer readFile.file.Close()
	for {
		readPrint := readFile.ReadPrint()
		//fmt.Println(readPrint)
		if readPrint != 3 {
			time.Sleep(time.Duration(200) * time.Microsecond)

		} else {
			time.Sleep(time.Duration(40) * time.Second)
		}
		if !follow {
			break
		}

	}
}
