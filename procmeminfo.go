package bt

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	memPath  = "/proc/meminfo"
	procPath = "/proc/self/status"
)

var (
	paths  = []string{memPath, procPath}
	mapper = map[string]string{
		"MemTotal":                   "system.memory.total",
		"MemFree":                    "system.memory.free",
		"MemAvailable":               "system.memory.available",
		"Buffers":                    "system.memory.buffers",
		"Cached":                     "system.memory.cached",
		"SwapCached":                 "system.memory.swap.cached",
		"Active":                     "system.memory.active",
		"Inactive":                   "system.memory.inactive",
		"SwapTotal":                  "system.memory.swap.total",
		"SwapFree":                   "system.memory.swap.free",
		"Dirty":                      "system.memory.dirty",
		"Writeback":                  "system.memory.writeback",
		"Slab":                       "system.memory.slab",
		"VmallocTotal":               "system.memory.vmalloc.total",
		"VmallocUsed":                "system.memory.vmalloc.used",
		"VmallocChunk":               "system.memory.vmalloc.chunk",
		"nonvoluntary_ctxt_switches": "sched.cs.involuntary",
		"voluntary_ctxt_switches":    "sched.cs.voluntary",
		"FDSize":                     "descriptor.count",
		"VmData":                     "vm.data.size",
		"VmLck":                      "vm.locked.size",
		"VmPTE":                      "vm.pte.size",
		"VmHWM":                      "vm.rss.peak",
		"VmRSS":                      "vm.rss.size",
		"VmLib":                      "vm.shared.size",
		"VmStk":                      "vm.stack.size",
		"VmSwap":                     "vm.swap.size",
		"VmPeak":                     "vm.vma.peak",
		"VmSize":                     "vm.vma.size",
	}
	memProcInfoAttrs   = map[string]string{}
	memProcInfoAttrsMu sync.RWMutex
)

func updateMemProcInfo() {
	for _, path := range paths {
		fileKVMap := readKeyValueFile(path)
		updateMemProcInfoAttrsFromFileKVMap(fileKVMap)
	}
}

func updateAttrsWithProcMemInfo(attributes map[string]interface{}) {
	memProcInfoAttrsMu.RLock()
	defer memProcInfoAttrsMu.RUnlock()

	for k, v := range memProcInfoAttrs {
		attributes[k] = v
	}
}

func updateMemProcInfoAttrsFromFileKVMap(fileKVMap map[string]string) {
	memProcInfoAttrsMu.Lock()
	defer memProcInfoAttrsMu.Unlock()
	for k, v := range fileKVMap {
		if attr, exists := mapper[k]; exists {
			memProcInfoAttrs[attr] = v
		}
	}
}

func readKeyValueFile(path string) map[string]string {
	file, err := os.Open(path)
	if err != nil {
		if Options.DebugBacktrace {
			log.Printf("readKeyValueFile err: %v", err)
		}
		return map[string]string{}
	}
	defer file.Close()
	return readKeyValueLines(file)
}

func readKeyValueLines(r io.Reader) map[string]string {
	m := make(map[string]string)
	reader := bufio.NewReader(r)
	for {
		l, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				if Options.DebugBacktrace {
					log.Printf("readKeyValueLines err: %v", err)
				}
				break
			}
		}

		values := strings.Split(string(l), ":")
		if len(values) == 2 {
			attr := values[0]
			value, err := getNormalizedValue(values[1])
			if err != nil {
				if Options.DebugBacktrace {
					log.Printf("readKeyValueLines err: %v", err)
				}
				continue
			}

			m[attr] = value
		}
	}
	return m
}

func getNormalizedValue(value string) (string, error) {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "kB") {
		value = strings.TrimSuffix(value, " kB")

		atoi, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return "", err
		}
		atoi *= 1024
		return fmt.Sprintf("%d", atoi), err
	}

	return value, nil
}
