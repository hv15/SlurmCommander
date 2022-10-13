package command

import (
	"encoding/json"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pja237/slurmcommander/internal/slurm"
)

// Calls `sinfo` to get node information for Cluster Tab
func GetSinfo(t time.Time) tea.Msg {
	var siJson slurm.SinfoJSON

	out, err := exec.Command(sinfoCmd, sinfoCmdSwitches...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error exec sinfo: %q\n", err)
	}

	err = json.Unmarshal(out, &siJson)
	if err != nil {
		log.Fatalf("Error unmarshall: %q\n", err)
	}

	return siJson

}

func TimedGetSinfo() tea.Cmd {
	// TODO: make timers configurable
	return tea.Tick(tick*time.Second, GetSinfo)
}

func QuickGetSinfo() tea.Cmd {
	// TODO: make timers configurable
	return tea.Tick(0*time.Second, GetSinfo)
}

// Calls `squeue` to get job information for Jobs Tab
func GetSqueue(t time.Time) tea.Msg {

	var sqJson slurm.SqueueJSON

	out, err := exec.Command(squeueCmd, squeueCmdSwitches...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error exec squeue: %q\n", err)
	}

	err = json.Unmarshal(out, &sqJson)
	if err != nil {
		log.Fatalf("Error unmarshall: %q\n", err)
	}

	return sqJson
}

func TimedGetSqueue() tea.Cmd {
	// TODO: make timers configurable
	return tea.Tick(tick*time.Second, GetSqueue)
}

func QuickGetSqueue() tea.Cmd {
	// TODO: make timers configurable
	return tea.Tick(0*time.Second, GetSqueue)
}

// Calls `sacct` to get job information for Jobs History Tab
func GetSacct(t time.Time) tea.Msg {

	var (
		sacct slurm.SacctList
	)

	re, err := regexp.Compile(`\D+`)
	if err != nil {
		// TODO: what do do now? return nil msg?
		return nil
	}

	// run sacct to get list of last 7 days
	out, err := exec.Command(sacctCmd, sacctCmdSwitches...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error exec sacct: %q\n", err)
	}

	// split output into lines
	for _, v := range strings.Split(string(out), "\n") {
		if v == "" {
			continue
		}
		// Test each line if jobid is num-only and append those lines to SacctList
		tmp := strings.Split(v, "|")
		if !re.MatchString(tmp[0]) {
			sacct = append(sacct, tmp)
		}
	}

	return sacct
}

func TimedGetSacct() tea.Cmd {
	return tea.Tick(tick*time.Second, GetSacct)
}

func QuickGetSacct() tea.Cmd {
	return tea.Tick(0*time.Second, GetSacct)
}

func SingleJobGetSacct(jobid string) tea.Cmd {
	return func() tea.Msg {
		var sacctJob slurm.SacctJob

		switches := append(sacctJobCmdSwitches, jobid)
		out, err := exec.Command(sacctJobCmd, switches...).CombinedOutput()
		if err != nil {
			log.Fatalf("Error exec sacct: %q\n", err)
		}

		err = json.Unmarshal(out, &sacctJob)
		if err != nil {
			log.Fatalf("Error unmarshall: %q\n", err)
		}

		return sacctJob
	}
}

type ScancelSent struct {
	Jobid string
}

func CallScancel(jobid string, l *log.Logger) tea.Cmd {
	return func() tea.Msg {
		var scret ScancelSent = ScancelSent{
			Jobid: jobid,
		}
		switches := append(scancelJobCmdSwitches, jobid)

		l.Printf("EXEC: %q %q\n", scancelJobCmd, switches)
		out, err := exec.Command(scancelJobCmd, switches...).CombinedOutput()
		if err != nil {
			l.Fatalf("Error exec scancel: %q\n", err)
		}
		l.Printf("EXEC output: %q\n", out)

		return scret
	}
}

type SHoldSent struct {
	Jobid string
}

func CallScontrolHold(jobid string, l *log.Logger) tea.Cmd {
	return func() tea.Msg {
		var scret SHoldSent = SHoldSent{
			Jobid: jobid,
		}
		switches := append(sholdJobCmdSwitches, jobid)

		l.Printf("EXEC: %q %q\n", sholdJobCmd, switches)
		out, err := exec.Command(sholdJobCmd, switches...).CombinedOutput()
		if err != nil {
			l.Fatalf("Error exec hold: %q\n", err)
		}
		l.Printf("EXEC output: %q\n", out)

		return scret
	}
}

type SRequeueSent struct {
	Jobid string
}

// TODO: unify this to a single function
func CallScontrolRequeue(jobid string, l *log.Logger) tea.Cmd {
	return func() tea.Msg {
		var scret SRequeueSent = SRequeueSent{
			Jobid: jobid,
		}
		switches := append(srequeueJobCmdSwitches, jobid)

		l.Printf("EXEC: %q %q\n", srequeueJobCmd, switches)
		out, err := exec.Command(srequeueJobCmd, switches...).CombinedOutput()
		if err != nil {
			l.Fatalf("Error exec requeue: %q\n", err)
		}
		l.Printf("EXEC output: %q\n", out)

		return scret
	}
}

type SBatchSent struct {
	JobFile string
}

// TODO: unify this to a single function
func CallSbatch(jobfile string, l *log.Logger) tea.Cmd {
	return func() tea.Msg {
		var scret SBatchSent = SBatchSent{
			JobFile: jobfile,
		}
		switches := append(sbatchCmdSwitches, jobfile)

		l.Printf("EXEC: %q %q\n", sbatchCmd, switches)
		out, err := exec.Command(sbatchCmd, switches...).CombinedOutput()
		if err != nil {
			l.Fatalf("Error exec sbatch: %q\n", err)
		}
		l.Printf("EXEC output: %q\n", out)

		return scret
	}
}
