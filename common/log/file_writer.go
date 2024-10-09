package log

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path"
	"time"
)

var pathVariableTable map[byte]func(*time.Time) int

type FileWriter struct {
	logLevelFloor int //用于控制日志的输出级别范围，确保只有符合级别的日志才会被写入文件。
	logLevelCeil  int
	filename      string
	pathFmt       string                 //日志文件的路径模式，用于文件轮转时生成新的文件名
	file          *os.File               //实际的文件指针和带缓存的文件写入器
	fileBufWriter *bufio.Writer          //用于写入日志数据
	actions       []func(*time.Time) int //存储用于轮转文件的时间变量函数（如按年、月、日等轮转）
	variables     []interface{}          //保存当前轮转条件（例如当前日志文件是基于哪一天创建的）。
}

// NewFileWriter
// 使用场景：
// 日志文件轮转：当日志文件按天、月或小时等单位轮转时，自动切换文件，防止单个日志文件过大。
// 按级别写日志：根据日志级别过滤不必要的日志，减轻日志文件的存储压力。
// 缓存写入：日志内容通过缓冲区写入文件，以提高性能，并可通过 Flush 立即将内容写入磁盘。
func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

func (w *FileWriter) Init() error {
	return w.CreateLogFile()
}

func (w *FileWriter) SetFileName(filename string) {
	w.filename = filename
}

func (w *FileWriter) SetLogLevelFloor(floor int) {
	w.logLevelFloor = floor
}

func (w *FileWriter) SetLogLevelCeil(ceil int) {
	w.logLevelCeil = ceil
}

// SetPathPattern
// 用于设置日志文件的路径模式，支持根据时间动态生成日志文件名。
// 它解析传入的模式字符串，并提取出时间变量，以便在日志轮转时创建带有时间戳的文件名。
// pattern = "./logs/gin_demo.wf.log.%Y%M%D%H"
func (w *FileWriter) SetPathPattern(pattern string) error {
	n := 0
	for _, c := range pattern {
		if c == '%' {
			n++
		}
	}

	if n == 0 {
		w.pathFmt = pattern
		return nil
	}

	w.actions = make([]func(*time.Time) int, 0, n)
	w.variables = make([]interface{}, n, n)
	tmp := []byte(pattern)

	variable := 0
	for _, c := range tmp {
		if variable == 1 {
			act, ok := pathVariableTable[c]
			if !ok {
				return errors.New("Invalid rotate pattern (" + pattern + ")")
			}
			w.actions = append(w.actions, act)
			variable = 0
			continue
		}
		if c == '%' {
			variable = 1
		}
	}

	for i, act := range w.actions {
		now := time.Now()
		w.variables[i] = act(&now)
	}
	//fmt.Printf("%v\n", w.variables)

	w.pathFmt = convertPatternToFmt(tmp)

	return nil
}

// Write 方法将符合级别范围的日志写入文件。它会通过 fileBufWriter 缓存写入，以提高写入性能。
func (w *FileWriter) Write(r *Record) error {
	if r.level < w.logLevelFloor || r.level > w.logLevelCeil {
		return nil
	}
	if w.fileBufWriter == nil {
		return errors.New("no opened file")
	}
	if _, err := w.fileBufWriter.WriteString(r.String()); err != nil {
		return err
	}
	return nil
}

// CreateLogFile 方法用于创建日志文件，保证在写入日志前，
// 日志文件存在且可写入。它还确保了日志文件的路径存在。
func (w *FileWriter) CreateLogFile() error {
	if err := os.MkdirAll(path.Dir(w.filename), 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if file, err := os.OpenFile(w.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err != nil {
		return err
	} else {
		w.file = file
	}

	if w.fileBufWriter = bufio.NewWriterSize(w.file, 8192); w.fileBufWriter == nil {
		return errors.New("new fileBufWriter failed.")
	}

	return nil
}

// Rotate 日志轮转用于定期切换日志文件（例如按日期、时间等）。当检测到轮转条件（如日期改变）满足时，
// 当前日志文件会被重命名（通过路径模式），并关闭旧文件，创建新文件继续写入。
func (w *FileWriter) Rotate() error {

	now := time.Now()
	v := 0
	rotate := false
	old_variables := make([]interface{}, len(w.variables))
	copy(old_variables, w.variables)

	//将now时间按年月日时进行拆分，然后之前的时间的年月日时进行拆分比较
	//如果不相同，则进行轮换；相当于每小时创建新的日志文件
	for i, act := range w.actions {
		v = act(&now)
		if v != w.variables[i] {
			w.variables[i] = v
			rotate = true
		}
	}
	//fmt.Printf("%v\n", w.variables)

	if rotate == false {
		return nil
	}

	if w.fileBufWriter != nil {
		if err := w.fileBufWriter.Flush(); err != nil {
			return err
		}
	}

	if w.file != nil {
		// 将文件以pattern形式改名并关闭
		filePath := fmt.Sprintf(w.pathFmt, old_variables...)

		if err := os.Rename(w.filename, filePath); err != nil {
			return err
		}

		if err := w.file.Close(); err != nil {
			return err
		}
	}

	return w.CreateLogFile()
}

// Flush 方法用于强制将缓存中的日志内容写入到磁盘。这可以确保在程序异常退出时，日志不会丢失。
func (w *FileWriter) Flush() error {
	if w.fileBufWriter != nil {
		return w.fileBufWriter.Flush()
	}
	return nil
}

func getYear(now *time.Time) int {
	return now.Year()
}

func getMonth(now *time.Time) int {
	return int(now.Month())
}

func getDay(now *time.Time) int {
	return now.Day()
}

func getHour(now *time.Time) int {
	return now.Hour()
}

func getMin(now *time.Time) int {
	return now.Minute()
}

// convertPatternToFmt 方法将路径中的时间变量替换为对应的格式化字符串，便于后续生成具体的日志文件名。
func convertPatternToFmt(pattern []byte) string {
	pattern = bytes.Replace(pattern, []byte("%Y"), []byte("%d"), -1)
	pattern = bytes.Replace(pattern, []byte("%M"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%D"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%H"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%m"), []byte("%02d"), -1)
	return string(pattern)
}

func init() {
	pathVariableTable = make(map[byte]func(*time.Time) int, 5)
	pathVariableTable['Y'] = getYear
	pathVariableTable['M'] = getMonth
	pathVariableTable['D'] = getDay
	pathVariableTable['H'] = getHour
	pathVariableTable['m'] = getMin
}
