package ymlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const DEFAULT_LOG_BUFFER_LENGTH = 2048

//const DEFAULT_Rotate_Duration = time.Hour * 24

var line = []byte("\n")

type RotateType int

const (
	RotateMinute RotateType = 1
	RotateHour   RotateType = 2
	RotateDay    RotateType = 3
)

type FileLoggerWriter struct {
	msgChan        chan []byte
	maxsizeCurSize int64
	maxRotateAge   time.Time

	file *os.File
	mu   sync.Mutex

	FileName     string
	realFileName string

	rotateTime     bool //
	rotateDuration time.Duration

	RotateType      RotateType
	MaxSizeByteSize int64 //byte

	//buffer
	ChanBufferLength int
	WriteFileBuffer  int
}

func (w *FileLoggerWriter) start() {
	if w == nil {
		panic("log.FileLoggerWriter error")
	}

	if w.RotateType == 0 {

	} else {
		switch w.RotateType {
		case RotateMinute:
			w.rotateDuration = time.Second * 60
			currentTime := time.Now()
			truncatedTime := currentTime.Truncate(time.Minute)
			w.maxRotateAge = truncatedTime.Add(w.rotateDuration)
		case RotateHour:
			w.rotateDuration = time.Hour
			currentTime := time.Now()
			truncatedTime := currentTime.Truncate(time.Hour)
			w.maxRotateAge = truncatedTime.Add(w.rotateDuration)
		case RotateDay:
			w.rotateDuration = time.Hour * 24
			currentTime := time.Now()
			truncatedTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
			w.maxRotateAge = truncatedTime.Add(w.rotateDuration)
		}
		w.rotateTime = true
	}

	//ChanBufferLength
	if w.msgChan == nil {
		if w.ChanBufferLength > 0 {
			w.msgChan = make(chan []byte, w.ChanBufferLength)
		} else {
			w.msgChan = make(chan []byte, DEFAULT_LOG_BUFFER_LENGTH)
		}
	}

	//
	if w.WriteFileBuffer == 0 {
		w.WriteFileBuffer = 1
	}

	w.fileRotate(true)

	// check 100 millisecond
	go func() {
		for range time.Tick(time.Microsecond * 100) {
			if err := w.fileRotate(false); err != nil {
				//fmt.Println(fmt.Printf("FileLogWriter(%s): %s\n", w.realFileName, err.Error()))
				fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.FileName, err)
				return
			}
		}
	}()

	go func() {
		defer func() {
			if w.file != nil {
				w.file.Close()
			}
		}()

		var msgBody []byte
		//https://blog.drkaka.com/batch-get-from-golangs-buffered-channel-9638573f0c6e 批量保存日志
		for {
			msgBody = msgBody[:0]

			//var items [][]byte
			msgBody = append(msgBody, <-w.msgChan...)
			msgBody = append(msgBody, line...)
			//items = append(items, <-w.msgChan)
			//Batch to obtain
		Remaining:
			for i := 1; i < w.WriteFileBuffer; i++ {
				select {
				case item := <-w.msgChan:
					//items = append(items, item)
					msgBody = append(msgBody, item...)
					msgBody = append(msgBody, line...)
				default:
					break Remaining
				}
			}
			//msgBody := join(items, line)
			n, err := w.file.Write(msgBody)
			if err != nil {
				fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.FileName, err)
				return
			}
			w.maxsizeCurSize += int64(n)

			// check fileRotate
			if err := w.fileRotate(false); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.FileName, err)
				return
			}
		}
	}()
}

// This is the FileLogWriter's output method
func (w *FileLoggerWriter) writeLog(msg []byte) {
	w.msgChan <- msg
}

func (w *FileLoggerWriter) checkRotate(now time.Time) bool {
	//check size
	if w.MaxSizeByteSize > 0 && w.maxsizeCurSize >= w.MaxSizeByteSize {
		return true
	}

	if w.maxRotateAge.Before(now) {
		return true
	}

	return false
}

func (w *FileLoggerWriter) close() {
	close(w.msgChan)
	w.file.Sync()
}

/*
*
Real file cutting, cutting method lock, and do secondary authentication
*/
func (w *FileLoggerWriter) fileRotate(init bool) error {
	//check un
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	if init {
		w.realFileName = getActualPathReplacePattern(w.FileName)
		// Open the log file
		checkDir(w.realFileName)
		fileInfo, err := os.Lstat(w.realFileName)
		if err == nil {
			w.maxsizeCurSize = fileInfo.Size()

			switch w.RotateType {
			case RotateMinute:
				w.rotateDuration = time.Second * 60
				truncatedTime := now.Truncate(time.Minute)
				w.maxRotateAge = truncatedTime.Add(w.rotateDuration)
			case RotateHour:
				w.rotateDuration = time.Hour
				truncatedTime := now.Truncate(time.Hour)
				w.maxRotateAge = truncatedTime.Add(w.rotateDuration)
			case RotateDay:
				w.rotateDuration = time.Hour * 24
				truncatedTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				w.maxRotateAge = truncatedTime.Add(w.rotateDuration)
			}

			w.rotateTime = true
		}
		fmt.Println(fmt.Sprintf("%s RotateLog>init>name>%s,Hour>%d", now.Format(time.RFC3339), w.realFileName, now.Hour()))
		fd, err := os.OpenFile(w.realFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
		if err == nil {
			w.file = fd
		} else {
			//fmt.Println(err)
			return fmt.Errorf("Rotate: %s\n", err)
		}
	}

	if w.RotateType > 0 && w.checkRotate(now) {

		fmt.Println(fmt.Sprintf("%s RotateLog>rotated>name>%s,Hour>%d", now.Format(time.RFC3339), w.realFileName, now.Hour()))

		w.realFileName = getActualPathReplacePattern(w.FileName)
		// Open the log file
		checkDir(w.realFileName)

		_, err := os.Lstat(w.realFileName)
		if err == nil {
			num := 1
			for ; ; num++ {
				fname := w.realFileName + fmt.Sprintf(".%d", num)
				_, err := os.Lstat(fname)
				if err != nil {
					err = os.Rename(w.realFileName, fname)
					break
				}
			}
		}

		var oldFd = w.file

		// 写入新文件
		newFd, err := os.OpenFile(w.realFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
		if err != nil {
			//fmt.Println(err)
			return err
		}
		w.file = newFd
		w.maxRotateAge = now.Add(w.rotateDuration)
		w.maxsizeCurSize = 0

		// 关闭老文件
		if w.maxsizeCurSize > 0 {
			err := oldFd.Sync()
			if err != nil {
				fmt.Println(err)
			}
			err = oldFd.Close()
			if err != nil {
				fmt.Println(err)
			}
		}

	} else {
		if w.maxsizeCurSize > 0 {
			w.file.Sync()
		}
	}
	return nil
}

func checkDir(fileName string) {
	dir := filepath.Dir(fileName)
	os.MkdirAll(dir, os.ModePerm)
}

func getActualPathReplacePattern(pattern string) string {
	now := time.Now()
	Y := fmt.Sprintf("%d", now.Year())
	M := fmt.Sprintf("%02d", now.Month())
	D := fmt.Sprintf("%02d", now.Day())
	H := fmt.Sprintf("%02d", now.Hour())
	m := fmt.Sprintf("%02d", now.Minute())
	s := fmt.Sprintf("%02d", now.Second())

	pattern = strings.Replace(pattern, "%Y", Y, -1)
	pattern = strings.Replace(pattern, "%M", M, -1)
	pattern = strings.Replace(pattern, "%D", D, -1)
	pattern = strings.Replace(pattern, "%H", H, -1)
	pattern = strings.Replace(pattern, "%m", m, -1)
	pattern = strings.Replace(pattern, "%s", s, -1)
	return pattern
}
