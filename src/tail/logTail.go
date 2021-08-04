package tail

import (
	"flag"
	"fmt"
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
			fmt.Print(string(data[:n]))
		} else {
			index := indexs[lines-line]
			fmt.Print(string(data[index+1 : n]))
		}

	case io.EOF:
	default:
		fmt.Println(err)
		return
	}
}

var (
	follow bool
	line   int
)

func init() {
	flag.BoolVar(&follow, "f", false, "on time")
	flag.IntVar(&line, "n", 10, "instead of the last 10")
}

func stream(path string) {
	flag.Parse()
	var err error
	var readFile ReadFile
	if len(os.Args) < 1 {
		flag.Usage()
		return
	}

	readFile.file, err = os.Open(path)
	//readFile.file, err = os.Open(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer readFile.file.Close()

	for {
		readFile.ReadPrint()
		if !follow {
			break
		}
	}
}
