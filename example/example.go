package main

import (
	"github.com/mok42/ymlog"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	///file log
	logger := ymlog.NewLogger(&ymlog.FileLoggerWriter{
		FileName:         "./logs/error.log",
		MaxSizeByteSize:  1024 * 1024 * 1024 * 1024 * 10,
		ChanBufferLength: 1024,
		WriteFileBuffer:  1024,
	},
	)
	logger2 := ymlog.NewLogger(&ymlog.FileLoggerWriter{
		FileName:         "./logs/error_%Y%M%D_%H%m%s.log",
		MaxSizeByteSize:  10 * 1024 * 1024 * 1024,
		RotateType:       ymlog.RotateMinute,
		ChanBufferLength: 1024,
		WriteFileBuffer:  100,
	},
	)
	logger.InfoString("init NewLogger log")

	///console
	logger1 := ymlog.NewLogger(&ymlog.OutLoggerWriter{
		Out: os.Stdout,
	})
	logger1.InfoString("init ConsoleLoggerWriter log")

	go echo1(logger2)
	go echo2(logger2)
	go echo3(logger2)
	go echo4(logger2)
	go echo5(logger2)
	go echo6(logger2)

	for {
		logger2.InfoString(time.Now().String() + strings.Repeat(" www.0000000.com ", rand.Intn(1000)))
	}

}

func echo1(logger2 *ymlog.Logger) {
	for {
		logger2.InfoString(time.Now().String() + strings.Repeat(" www.11111111.com ", rand.Intn(3000)))
	}
}

func echo2(logger2 *ymlog.Logger) {
	for {
		logger2.InfoString(time.Now().String() + strings.Repeat(" www.22222222.com ", rand.Intn(3000)))
	}
}

func echo3(logger2 *ymlog.Logger) {
	for {
		logger2.InfoString(time.Now().String() + strings.Repeat(" www.33333333.com ", rand.Intn(3000)))
	}
}

func echo4(logger2 *ymlog.Logger) {
	for {
		logger2.InfoString(time.Now().String() + strings.Repeat(" www.44444444.com ", rand.Intn(4000)))
	}
}

func echo5(logger2 *ymlog.Logger) {
	for {
		logger2.InfoString(time.Now().String() + strings.Repeat(" www.55555555.com ", rand.Intn(5000)))
	}
}

func echo6(logger2 *ymlog.Logger) {
	for {
		logger2.InfoString(time.Now().String() + strings.Repeat(" www.66666666.com ", rand.Intn(6000)))
	}
}
