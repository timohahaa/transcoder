package encoder

import (
	"os"
	"path/filepath"
	"runtime"
)

type (
	Config struct {
		ComposerAddrs     []string `arg:"required,-,--,env:COMPOSER_ADDRS"`
		CPUQuota          int      `arg:"-,--,env:CPU_QUOTA"`
		WorkDir           string   `arg:"-,--,env:WORK_DIR"`
		MaxTasksPerWorker int      `arg:"-,--,env:MAX_TASKS_PER_WORKER"`
	}
)

func (c *Config) setDefaults() {
	if c.CPUQuota <= 0 {
		c.CPUQuota = runtime.NumCPU() - 1 // leave 1 cpu to go process
	}
	if c.WorkDir == "" {
		c.WorkDir = filepath.Join(os.TempDir(), "encoder")
	}
	if c.MaxTasksPerWorker <= 0 {
		c.MaxTasksPerWorker = 1
	}
}
