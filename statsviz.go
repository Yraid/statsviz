package statsviz

import (
	"os"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
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

func pidUsageDuringTime(info *pidInfo) {
	sysInfo, err := pidusage.GetStat(os.Getpid())
	if err == nil {
		info.CPUPercentage = float64(sysInfo.CPU / 100.0)
		info.MemBytes = uint64(sysInfo.Memory)
	}
}

// sendStats indefinitely send runtime statistics on the websocket connection.
func sendStats(conn *websocket.Conn, frequency time.Duration) error {
	tick := time.NewTicker(frequency)
	defer tick.Stop()

	stats := stats{GoVersion: runtime.Version()}
	for range tick.C {
		runtime.ReadMemStats(&stats.Mem)
		//sysProcessInfo(&stats.ProcessInfo)
		pidUsageDuringTime(&stats.ProcessInfo)
		stats.NumGoroutine = runtime.NumGoroutine()
		if err := conn.WriteJSON(stats); err != nil {
			return err
		}
	}

	panic("unreachable")
}
