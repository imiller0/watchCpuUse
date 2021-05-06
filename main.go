package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/procfs"
	"os"
	"strconv"
	"time"
	"strings"
)

type Options struct {
	count *int
	interval *int
	verbose *bool
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [opts] <pid>\n", os.Args[0]);
	flag.PrintDefaults()
}

func main() {
	var opts Options

	flag.Usage = usage
	opts.count = flag.Int("count", 1, "(optional) Number of sample intervals to monitor. 0 runs forever")
	opts.interval = flag.Int("interval", 10, "(optional) sample interval in seconds")
	opts.verbose = flag.Bool("verbose", false, "(optional) Verbose output")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Try the argument as a pid first
	ppid_x, err := strconv.ParseInt(flag.Arg(0), 10, 32)
	if err != nil {
		// Lookup the process by string key
		ppid_lookup,err := lookupProcess(flag.Arg(0))
		if err != nil {
			fmt.Println("Invalid parent pid: ", err)
			os.Exit(1)
		}
		ppid_x = int64(ppid_lookup)
	}
	ppid := int(ppid_x)

	// Get a handle to the parent process before monitoring
	parent,err := procfs.NewProc(ppid)
	if err != nil {
		fmt.Printf("Could not monitor pid: %d\n", ppid)
		os.Exit(1)
	}
	fmt.Printf("Monitoring pid %d \"%s\"\n", ppid, showCmdline("",parent))
	for idx := 0; idx < *opts.count || *opts.count == 0; idx++ {
		err := runSample(opts, ppid)
		if err != nil {
			fmt.Println("ERROR: ",err)
			os.Exit(1)
		}
	}
}

func runSample (opts Options, ppid int) error {
	// tick is usually 100. Binding in C support makes this
	// program less portable so using a fixed value here for now.
	tck := 100

	startProcs, err := getRelevantProcs(ppid)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(*opts.interval) * time.Second)
	endProcs, err := getRelevantProcs(ppid)
	if err != nil {
		return err
	}

	var totalUtime, totalStime uint = 0,0
	for key, end := range endProcs {
		if start, ok := startProcs[key]; ok {
			if *opts.verbose {
				fmt.Printf("found %d (%s) %d - %d\n", key, start.Comm, end.UTime, start.UTime)
			}
			totalUtime += (end.UTime - start.UTime)
			totalStime += (end.STime - start.STime)
			delete(startProcs, key)
		} else {
			fmt.Printf("new process appeared %d (%s)\n", key, end.Comm)
			totalUtime += (end.UTime)
			totalStime += (end.STime)
		}
	}
	for key, val := range startProcs {
		fmt.Printf("ended %d (%s)\n", key, val.Comm)
	}
	usec := float64(totalUtime)/float64(tck)
	ssec := float64(totalStime)/float64(tck)
	fmt.Printf("Total %0.3f sec (%.3f %%cpu) : %0.3f user, %0.3f sys ( %.3f %% user, %.3f %% sys ) @ %s\n",
		usec + ssec,
		(usec + ssec) * 100 / float64(*opts.interval),
		usec, ssec,
		100 * usec / float64(*opts.interval),
		100 * ssec / float64(*opts.interval),
		time.Now().Format(time.RFC1123Z),
	)

	return nil
}

func getRelevantProcs(ppid int) (map[int]procfs.ProcStat, error) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}
	procs := make(map[int]procfs.ProcStat)
	pending := make([]int, 1)
	pending[0] = ppid
	for len(pending) > 0 {
		curPid := pending[0]
		pending = pending[1:]
		cur, _ := fs.Proc(curPid)
		curStat, _ := cur.NewStat()
		procs[curPid] = curStat // Add to the set of procs we've found
		allProcs, _ := fs.AllProcs()
		for _, proc := range allProcs {
			stat, _ := proc.NewStat()
			//fmt.Printf("Compare %d %d/%d %d/%d\n",stat.PID, stat.PPID, curStat.PID, stat.PGRP, curStat.PGRP)
			if stat.PPID == curStat.PID {
				//fmt.Printf("%d\n", proc.PID)
				if _, ok := procs[stat.PID]; !ok {
					procs[stat.PID] = stat
					pending = append(pending, stat.PID)
				}
			}
		}
	}
	return procs, nil
}

func lookupProcess(key string) (int, error) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return 0, err
	}
	allProcs, _ := fs.AllProcs()
	matches := make([]procfs.Proc, 0)
	for _, proc := range allProcs {
		if proc.PID == os.Getpid() {
			continue
		}
		comm, _ := proc.Comm()
		if strings.Contains(comm, key) {
			matches = append(matches, proc)
			continue
		}
		cmdline, _ := proc.CmdLine()
		for _,term := range cmdline {
			if strings.Contains(term, key) {
				matches = append(matches, proc)
				break
			}
		}
	}

	var index int
	if len(matches) == 0 {
		fmt.Printf("No processes match \"%s\"\n", key)
		os.Exit(1)
	} else if len(matches) == 1 {
		index = 0
	} else {
		for id,entry := range matches {
			fmt.Printf("  [%2d]  %8d %s\n", id, entry.PID, showCmdline(key, entry) )
		}
		fmt.Print("Enter index of process to monitor: ")
		fmt.Scanf("%d", &index)
		if index < 0 || index >= len(matches) {
			fmt.Println("Index is out of range")
			os.Exit(1)
		}
	}

	return matches[index].PID, nil
}

func showCmdline(key string, proc procfs.Proc) (string) {
	comm,_ := proc.Comm()
	args,_ := proc.CmdLine()
	cmdline := strings.Join(args, " ")
	if len(cmdline) > 80 {
		idx := strings.Index(cmdline, key)
		if idx >= 0 {
			start := idx - 100
			pre_elipses := "..."
			if start < 0 {
				start = 0
				pre_elipses = "..."
			}
			end := idx + 100
			post_elipses := "..."
			if end > len(cmdline) {
				end = len(cmdline)
				post_elipses = ""
			}
			cmdline = pre_elipses + cmdline[start:end] + post_elipses
		} else {
			cmdline = cmdline[:77] + "..."
		}
	}

	return fmt.Sprintf("%s   %s", comm, cmdline)
}
