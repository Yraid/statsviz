package statsviz

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	ps "github.com/mitchellh/go-ps"
	"github.com/struCoder/pidusage"
)

type pidInfo struct {
	CPUPercentage float64
	MemBytes      uint64
}

type stats struct {
	GoVersion    string
	Mem          runtime.MemStats
	ProcessInfo  pidInfo
	NumGoroutine int
}

// return child process if it has a parent process
func getPid() int {
	ppid := os.Getppid()
	pid := os.Getpid()
	fname := filepath.Base(os.Args[0])

	parent, err := ps.FindProcess(ppid)
	if err != nil {
		return -1
	}
	if parent.Executable() == fname {
		// it's child
		return pid
	}

	plist, err := ps.Processes()
	if err != nil {
		return -1
	}
	// it might be a parent, looking for the child...
	for _, p := range plist {
		if p.Executable() == fname && p.Pid() != pid {
			// it's child
			return p.Pid()
		}
	}

	// it's a single process
	return pid
}

func pidUsageDuringTime(pid int, info *pidInfo) {
	sysInfo, err := pidusage.GetStat(pid)
	if err == nil {
		info.CPUPercentage = float64(sysInfo.CPU / 100.0)
		info.MemBytes = uint64(sysInfo.Memory)
	}
}

// sendStats indefinitely send runtime statistics on the websocket connection.
func sendStats(conn *websocket.Conn, frequency time.Duration) error {
	tick := time.NewTicker(frequency)
	defer tick.Stop()

	pid := getPid()
	stats := stats{GoVersion: runtime.Version()}
	for range tick.C {
		runtime.ReadMemStats(&stats.Mem)
		pidUsageDuringTime(pid, &stats.ProcessInfo)
		stats.NumGoroutine = runtime.NumGoroutine()
		if err := conn.WriteJSON(stats); err != nil {
			return err
		}
	}

	panic("unreachable")
}
