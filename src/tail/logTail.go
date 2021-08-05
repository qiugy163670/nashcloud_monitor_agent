package tail

import (
	log "github.com/cihub/seelog"
	"io"
	"os"
	"runtime"
	"strings"
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

			if strings.Index(string(data[:n]), "block") == -1 {
				logStr := strings.ReplaceAll(string(data[:n]), "\n", "")
				logs := Json2Struct(logStr)
				//mainLog := MainLogSync(logs)
				MainLogSync(logStr, logs.Time)
			}

			//logs := Json2Struct(string(data[:n]))
			//mainLog := MainLogSync(logs)
			//MainLogSync(logs)

		} else {

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
