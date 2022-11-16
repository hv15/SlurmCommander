package model

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/pja237/slurmcommander-dev/internal/generic"
	"github.com/pja237/slurmcommander-dev/internal/keybindings"
	"github.com/pja237/slurmcommander-dev/internal/styles"
	"github.com/pja237/slurmcommander-dev/internal/version"
)

// genTabs() generates top tabs
func (m Model) genTabs() string {

	var doc strings.Builder

	tlist := make([]string, len(tabs))
	for i, v := range tabs {
		if i == int(m.ActiveTab) {
			tlist = append(tlist, styles.TabActiveTab.Render(v))
		} else {
			tlist = append(tlist, styles.Tab.Render(v))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tlist...)

	//gap := tabGap.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(row)-2)))
	gap := styles.TabGap.Render(strings.Repeat(" ", max(0, m.winW-lipgloss.Width(row)-2)))
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	doc.WriteString(row + "\n")

	return doc.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) tabJobDetails() (scr string) {

	var (
		runT  time.Duration
		waitT time.Duration
	)

	// race between View() call and command.SingleJobGetSacct(m.JobDetailsTab.SelJobID) call
	switch {
	//case m.JobDetailsTab.SelJobID == "":
	//	return "Select a job from the Job History tab.\n"
	case m.JobDetailsTab.SelJobIDNew == -1:
		return "Select a job from the Job History tab.\n"
	//case len(m.SacctSingleJobHist.Jobs) == 0:
	//	return fmt.Sprintf("Waiting for job %s info...\n", m.JobDetailsTab.SelJobID)
	case len(m.JobHistTab.SacctHistFiltered.Jobs) == 0:
		//return fmt.Sprintf("Waiting for job %s info...\n", m.JobDetailsTab.SelJobID)
		return "Select a job from the Job History tab.\n"
	}

	//width := m.Globals.winW - 10

	//job := m.SacctSingleJobHist.Jobs[0]
	// NEW:
	job := m.JobHistTab.SacctHistFiltered.Jobs[m.JobDetailsTab.SelJobIDNew]

	m.Log.Printf("Job Details req %#v ,got: %#v\n", m.JobDetailsTab.SelJobID, job.JobId)

	// TODO: consider moving this to a viewport...

	fmtStr := "%-20s : %-60s\n"
	fmtStrX := "%-20s : %-60s"

	head := ""
	waitT = time.Unix(int64(*job.Time.Start), 0).Sub(time.Unix(int64(*job.Time.Submission), 0))
	// If job is RUNNING, use Elapsed instead of Sub (because End=0)
	if *job.State.Current == "RUNNING" {
		runT = time.Duration(int64(*job.Time.Elapsed) * int64(time.Second))
	} else {
		runT = time.Unix(int64(*job.Time.End), 0).Sub(time.Unix(int64(*job.Time.Start), 0))
	}

	head += styles.StatsSeparatorTitle.Render(fmt.Sprintf(fmtStrX, "Job ID", strconv.Itoa(*job.JobId)))
	head += "\n"
	head += fmt.Sprintf(fmtStr, "Job Name", *job.Name)
	head += fmt.Sprintf(fmtStr, "User", *job.User)
	head += fmt.Sprintf(fmtStr, "Group", *job.Group)
	head += fmt.Sprintf(fmtStr, "Job Account", *job.Account)
	head += fmt.Sprintf(fmtStr, "Job Submission", time.Unix(int64(*job.Time.Submission), 0).String())
	head += fmt.Sprintf(fmtStr, "Job Start", time.Unix(int64(*job.Time.Start), 0).String())
	// Running jobs have End==0
	if *job.State.Current == "RUNNING" {
		head += fmt.Sprintf(fmtStr, "Job End", "RUNNING")
	} else {
		head += fmt.Sprintf(fmtStr, "Job End", time.Unix(int64(*job.Time.End), 0).String())
	}
	head += fmt.Sprintf(fmtStr, "Job Wait time", waitT.String())
	head += fmt.Sprintf(fmtStr, "Job Run time", runT.String())
	head += fmt.Sprintf(fmtStr, "Partition", *job.Partition)
	head += fmt.Sprintf(fmtStr, "Priority", strconv.Itoa(*job.Priority))
	head += fmt.Sprintf(fmtStr, "QoS", *job.Qos)

	scr += styles.JobStepBoxStyle.Width(90).Render(head)
	scr += "\n"

	scr += styles.TextYellow.Render(fmt.Sprintf("Steps count: %d", len(*job.Steps)))

	steps := ""
	for i, v := range *job.Steps {

		m.Log.Printf("Job Details, step: %d name: %s\n", i, *v.Step.Name)
		step := styles.StatsSeparatorTitle.Render(fmt.Sprintf(fmtStrX, "Name", *v.Step.Name))
		step += "\n"
		step += fmt.Sprintf(fmtStr, "Nodes", *v.Nodes.Range)
		if *v.State != "COMPLETED" {
			step += styles.JobStepExitStatusRed.Render(fmt.Sprintf(fmtStrX, "State", *v.State))
			step += "\n"
		} else {
			//step += fmt.Sprintf(fmtStr, "State", *v.State)
			step += styles.JobStepExitStatusGreen.Render(fmt.Sprintf(fmtStrX, "State", *v.State))
			step += "\n"
		}
		if *v.ExitCode.Status != "SUCCESS" {
			step += styles.JobStepExitStatusRed.Render(fmt.Sprintf(fmtStrX, "ExitStatus", *v.ExitCode.Status))
			step += "\n"
		} else {
			step += styles.JobStepExitStatusGreen.Render(fmt.Sprintf(fmtStrX, "ExitStatus", *v.ExitCode.Status))
			step += "\n"
		}
		if *v.ExitCode.Status == "SIGNALED" {
			step += styles.JobStepExitStatusRed.Render(fmt.Sprintf(fmtStrX, "Signal ID", strconv.Itoa(*v.ExitCode.Signal.SignalId)))
			step += "\n"
			step += styles.JobStepExitStatusRed.Render(fmt.Sprintf(fmtStrX, "SignalName", *v.ExitCode.Signal.Name))
			step += "\n"
		}
		if v.KillRequestUser != nil {
			step += fmt.Sprintf(fmtStr, "KillReqUser", *v.KillRequestUser)
		}
		step += fmt.Sprintf(fmtStr, "Tasks", strconv.Itoa(*v.Tasks.Count))

		// TODO: TRES part needs quite some love...
		tres := ""
		tresAlloc := ""

		//tresReqMin := ""
		//tresReqMax := ""
		//tresReqAvg := ""
		//tresReqTotal := ""
		//tresConMax := ""
		//tresConMin := ""
		// TRES: allocated
		tresAlloc += "\nALLOCATED:\n"
		m.Log.Printf("Dumping step allocation: %#v\n", *v.Tres.Allocated)
		m.Log.Printf("ALLOCATED:\n")
		for i, t := range *v.Tres.Allocated {
			if t.Count != nil {
				m.Log.Printf("Dumping type %d : %s - %d\n", i, *t.Type, *t.Count)
				tresAlloc += "* "
				if *t.Type == "gres" {
					// TODO:
					//fmtStr := "%-20s : %-60s\n"
					tresAlloc += fmt.Sprintf(fmtStr, *t.Type, strings.Join([]string{*t.Name, strconv.Itoa(*t.Count)}, ":"))
				} else {
					// TODO:
					tresAlloc += fmt.Sprintf(fmtStr, *t.Type, strconv.Itoa(*t.Count))
				}
			}
		}
		//// REQUESTED:MIN
		//tresReqMin += "REQUESTED:Min:\n"
		//m.Log.Printf("REQ:Min\n")
		//for i, t := range *v.Tres.Requested.Min {
		//	if t.Count != nil {
		//		m.Log.Printf("Dumping type %d : %s - %d\n", i, *t.Type, *t.Count)
		//		tresReqMin += " "
		//		tresReqMin += fmt.Sprintf(fmtStr, *t.Type, strconv.Itoa(*t.Count))
		//	}
		//}
		//// REQUESTED:MAX
		//m.Log.Printf("REQ:Max\n")
		//tresReqMax += "REQUESTED:Max:\n"
		//for i, t := range *v.Tres.Requested.Min {
		//	if t.Count != nil {
		//		m.Log.Printf("Dumping type %d : %s - %d\n", i, *t.Type, *t.Count)
		//		tresReqMax += " "
		//		tresReqMax += fmt.Sprintf(fmtStr, *t.Type, strconv.Itoa(*t.Count))
		//	}
		//}
		//// REQUESTED:AVG
		//m.Log.Printf("REQ:Avg\n")
		//tresReqAvg += "REQUESTED:Avg:\n"
		//for i, t := range *v.Tres.Requested.Average {
		//	if t.Count != nil {
		//		m.Log.Printf("Dumping type %d : %s - %d\n", i, *t.Type, *t.Count)
		//		tresReqAvg += fmt.Sprintf(fmtStr, *t.Type, strconv.Itoa(*t.Count))
		//	}
		//}
		//// REQUESTED:TOT
		//tresReqAvg += "REQUESTED:Tot:\n"
		//m.Log.Printf("REQ:Tot\n")
		//for i, t := range *v.Tres.Requested.Total {
		//	if t.Count != nil {
		//		m.Log.Printf("Dumping type %d : %s - %d\n", i, *t.Type, *t.Count)
		//		tresReqTotal += fmt.Sprintf(fmtStr, *t.Type, strconv.Itoa(*t.Count))
		//	}
		//}
		//// Consumed:Min
		//tresConMin += "CONSUMED:Min:\n"
		//m.Log.Printf("CONS:Min\n")
		//for i, t := range *v.Tres.Consumed.Min {
		//	if t.Count != nil {
		//		m.Log.Printf("Dumping type %d : %s - %d\n", i, *t.Type, *t.Count)
		//		tresConMin += fmt.Sprintf(fmtStr, *t.Type, strconv.Itoa(*t.Count))
		//	}
		//}
		//// Consumed:Max
		//tresConMax += "CONSUMED:Max:\n"
		//m.Log.Printf("CONS:Max\n")
		//for i, t := range *v.Tres.Consumed.Max {
		//	if t.Count != nil {
		//		m.Log.Printf("Dumping type %d : %s - %d\n", i, *t.Type, *t.Count)
		//		tresConMax += fmt.Sprintf(fmtStr, *t.Type, strconv.Itoa(*t.Count))
		//	}
		//}
		//tres = lipgloss.JoinHorizontal(lipgloss.Top, styles.TresBox.Render(tresAlloc), styles.TresBox.Width(40).Render(tresConMax))

		// For now, show just allocated, later rework this whole part
		tres = styles.TresBox.Render(tresAlloc)

		step += tres

		// when the step is finished, append it to steps string
		steps += "\n" + styles.JobStepBoxStyle.Render(step)
	}
	scr += steps

	return scr
}

func (m Model) tabJobFromTemplate() string {

	if m.EditTemplate {
		return m.TemplateEditor.View()
	} else {
		if len(m.JobFromTemplateTab.TemplatesList) == 0 {
			return styles.NotFound.Render("\nNo templates found!\n")
		} else {
			return m.TemplatesTable.View()
		}
	}
}

func (m Model) tabClusterBars() string {
	var (
		scr     string = ""
		cpuPerc float64
		memPerc float64
	)

	sel := m.SinfoTable.Cursor()
	m.Log.Printf("ClusterTab Selected: %d\n", sel)
	m.Log.Printf("ClusterTab len results: %d\n", len(m.JobClusterTab.SinfoFiltered.Nodes))
	m.JobClusterTab.CpuBar = progress.New(progress.WithGradient("#277BC0", "#FFCB42"))
	m.JobClusterTab.MemBar = progress.New(progress.WithGradient("#277BC0", "#FFCB42"))
	if len(m.JobClusterTab.SinfoFiltered.Nodes) > 0 && sel != -1 {
		cpuPerc = float64(*m.JobClusterTab.SinfoFiltered.Nodes[sel].AllocCpus) / float64(*m.JobClusterTab.SinfoFiltered.Nodes[sel].Cpus)
		memPerc = float64(*m.JobClusterTab.SinfoFiltered.Nodes[sel].AllocMemory) / float64(*m.JobClusterTab.SinfoFiltered.Nodes[sel].RealMemory)

		scr += fmt.Sprintf("CPU used/total: %d/%d\n", *m.JobClusterTab.SinfoFiltered.Nodes[sel].AllocCpus, *m.JobClusterTab.SinfoFiltered.Nodes[sel].Cpus)
		scr += m.CpuBar.ViewAs(cpuPerc)
		scr += "\n"
		scr += fmt.Sprintf("MEM used/total: %d/%d\n", *m.JobClusterTab.SinfoFiltered.Nodes[sel].AllocMemory, *m.JobClusterTab.SinfoFiltered.Nodes[sel].RealMemory)
		scr += m.MemBar.ViewAs(memPerc)
		scr += "\n\n"
	} else {
		cpuPerc = 0
		memPerc = 0
		scr += fmt.Sprintf("CPU used/total: %d/%d\n", 0, 0)
		scr += m.CpuBar.ViewAs(cpuPerc)
		scr += "\n"
		scr += fmt.Sprintf("MEM used/total: %d/%d\n", 0, 0)
		scr += m.MemBar.ViewAs(memPerc)
		scr += "\n\n"

	}

	return scr
}
func (m Model) tabCluster() string {

	scr := m.SinfoTable.View() + "\n"

	return scr
}

func (m Model) tabAbout() string {

	s := "Version: " + version.BuildVersion + "\n"
	s += "Commit : " + version.BuildCommit + "\n"

	s += `
petar.jager@imba.oeaw.ac.at
CLIP-HPC Team @ VBC

Contributors:
`

	return s
}

func (m Model) getClusterCounts() string {
	var (
		ret string
		cpp string
		mpp string
		nps string
	)

	fmtStrCpu := "%-10s : %4d / %4d %2.0f%%\n"
	fmtStrMem := "%-10s : %8d / %8d %2.0f%%\n"
	fmtStrNPS := "%-15s : %4d\n"
	fmtTitle := "%-40s"

	cpp += styles.TextYellowOnBlue.Render(fmt.Sprintf(fmtTitle, "CPUs per Partition (used/total)"))
	cpp += "\n"
	for _, v := range m.JobClusterTab.Breakdowns.CpuPerPart {
		cpp += fmt.Sprintf(fmtStrCpu, v.Name, v.Count, v.Total, float32(v.Count)/float32(v.Total)*100)
	}

	mpp += styles.TextYellowOnBlue.Render(fmt.Sprintf(fmtTitle, "Mem per Partition (used/total)"))
	mpp += "\n"
	for _, v := range m.JobClusterTab.Breakdowns.MemPerPart {
		mpp += fmt.Sprintf(fmtStrMem, v.Name, v.Count, v.Total, float32(v.Count)/float32(v.Total)*100)
	}

	nps += styles.TextYellowOnBlue.Render(fmt.Sprintf(fmtTitle, "Nodes per State"))
	nps += "\n"
	for _, v := range m.JobClusterTab.Breakdowns.NodesPerState {
		nps += fmt.Sprintf(fmtStrNPS, v.Name, v.Count)
	}

	cpp = styles.CountsBox.Render(cpp)
	mpp = styles.CountsBox.Render(mpp)
	nps = styles.CountsBox.Render(nps)

	ret = lipgloss.JoinHorizontal(lipgloss.Top, cpp, mpp, nps)

	return ret
}

func (m *Model) genTabHelp() string {
	var th string
	switch m.ActiveTab {
	case tabJobs:
		th = "List of jobs in the queue"
	case tabJobHist:
		th = fmt.Sprintf("List of jobs in the last %d days from all user associated accounts. (timeout: %d seconds)", m.JobHistTab.JobHistStart, m.JobHistTab.JobHistTimeout)
	case tabJobDetails:
		th = "Job details, select a job from Job History tab"
	case tabJobFromTemplate:
		th = "Edit and submit one of the job templates"
	case tabCluster:
		th = "List and status of cluster nodes"
	default:
		th = "SlurmCommander"
	}
	return th + "\n"
}

// Generate statistics string, horizontal.
func GenCountStr(cnt map[string]uint, l *log.Logger) string {
	var (
		scr string
	)

	sm := make([]struct {
		name string
		val  uint
	}, 0)

	// place map to slice
	for k, v := range cnt {
		sm = append(sm, struct {
			name string
			val  uint
		}{name: k, val: uint(v)})
	}

	// sort it
	sort.Slice(sm, func(i, j int) bool {
		if sm[i].name < sm[j].name {
			return true
		} else {
			return false
		}
	})

	// print it out
	scr = "Count: "
	for _, v := range sm {
		scr += fmt.Sprintf("%s: %d ", v.name, v.val)
	}
	scr += "\n\n"

	return scr
}

func (m Model) JobClusterTabStats() string {
	var str string

	m.Log.Printf("JobClusterTabStats called\n")

	sel := m.JobClusterTab.SinfoTable.Cursor()
	str += styles.StatsSeparatorTitle.Render(fmt.Sprintf("%-30s", "Nodes states (filtered):"))
	str += "\n"

	if len(m.JobClusterTab.SinfoFiltered.Nodes) > 0 {
		//str += generic.GenCountStrVert(m.JobClusterTab.Stats.StateCnt, m.Log)
		str += generic.GenCountStrVert(m.JobClusterTab.Stats.StateSimpleCnt, m.Log)
	}

	str += styles.StatsSeparatorTitle.Render(fmt.Sprintf("%-30s", "Selected node:"))

	if len(m.JobClusterTab.SinfoFiltered.Nodes) > 0 && sel != -1 {
		str += "\n"
		str += fmt.Sprintf("%-15s: %s\n", "Arch", *m.JobClusterTab.SinfoFiltered.Nodes[sel].Architecture)
		str += fmt.Sprintf("%-15s: %s\n", "Features", *m.JobClusterTab.SinfoFiltered.Nodes[sel].ActiveFeatures)
		str += fmt.Sprintf("%-15s: %s\n", "TRES", *m.JobClusterTab.SinfoFiltered.Nodes[sel].Tres)
		if m.JobClusterTab.SinfoFiltered.Nodes[sel].TresUsed != nil {
			str += fmt.Sprintf("%-15s: %s\n", "TRES Used", *m.JobClusterTab.SinfoFiltered.Nodes[sel].TresUsed)
		} else {
			str += fmt.Sprintf("%-15s: %s\n", "TRES Used", "")
		}
		str += fmt.Sprintf("%-15s: %s\n", "GRES", *m.JobClusterTab.SinfoFiltered.Nodes[sel].Gres)
		str += fmt.Sprintf("%-15s: %s\n", "GRES Used", *m.JobClusterTab.SinfoFiltered.Nodes[sel].GresUsed)
		str += fmt.Sprintf("%-15s: %s\n", "Partitions", strings.Join(*m.JobClusterTab.SinfoFiltered.Nodes[sel].Partitions, ","))
	}
	return str
}

func (m Model) JobTabStats() string {

	m.Log.Printf("JobTabStats called\n")

	//str := "Queue statistics (filtered):\n\n"
	str := styles.StatsSeparatorTitle.Render(fmt.Sprintf("%-30s", "Job states (filtered):"))
	str += "\n\n"

	str += generic.GenCountStrVert(m.JobTab.Stats.StateCnt, m.Log)

	str += styles.StatsSeparatorTitle.Render(fmt.Sprintf("%-30s", "Pending jobs:"))
	str += "\n\n"
	str += fmt.Sprintf("%-10s : %s\n", " ", "dd-hh:mm:ss")
	str += fmt.Sprintf("%-10s : %s\n", "MinWait", generic.HumanizeDuration(m.JobTab.Stats.MinWait, m.Log))
	str += fmt.Sprintf("%-10s : %s\n", "AvgWait", generic.HumanizeDuration(m.JobTab.Stats.AvgWait, m.Log))
	str += fmt.Sprintf("%-10s : %s\n", "MedWait", generic.HumanizeDuration(m.JobTab.Stats.MedWait, m.Log))
	str += fmt.Sprintf("%-10s : %s\n", "MaxWait", generic.HumanizeDuration(m.JobTab.Stats.MaxWait, m.Log))

	str += "\n"
	str += styles.StatsSeparatorTitle.Render(fmt.Sprintf("%-30s", "Running jobs:"))
	str += "\n\n"
	str += fmt.Sprintf("%-10s : %s\n", " ", "dd-hh:mm:ss")
	str += fmt.Sprintf("%-10s : %s\n", "MinRun", generic.HumanizeDuration(m.JobTab.Stats.MinRun, m.Log))
	str += fmt.Sprintf("%-10s : %s\n", "AvgRun", generic.HumanizeDuration(m.JobTab.Stats.AvgRun, m.Log))
	str += fmt.Sprintf("%-10s : %s\n", "MedRun", generic.HumanizeDuration(m.JobTab.Stats.MedRun, m.Log))
	str += fmt.Sprintf("%-10s : %s\n", "MaxRun", generic.HumanizeDuration(m.JobTab.Stats.MaxRun, m.Log))

	return str
}

func (m Model) View() string {

	var (
		scr        strings.Builder
		MainWindow strings.Builder
	)

	// HEADER / TABS
	scr.WriteString(m.genTabs())
	scr.WriteString(m.genTabHelp())

	if m.Debug {
		// One debug line
		scr.WriteString(fmt.Sprintf("%s Width: %d Height: %d ErrorMsg: %s\n", styles.TextRed.Render("DEBUG ON:"), m.Globals.winW, m.Globals.winH, m.Globals.ErrorMsg))
	}

	if m.Globals.ErrorHelp != "" {
		scr.WriteString(styles.ErrorHelp.Render(fmt.Sprintf("ERROR: %s", m.Globals.ErrorHelp)))
		scr.WriteString("\n")
	}

	// PICK and RENDER ACTIVE TAB
	switch m.ActiveTab {
	case tabJobs:
		m.Log.Printf("CALL JobTab.View()\n")
		MainWindow.WriteString(m.JobTab.View(m.Log))

	case tabJobHist:
		m.Log.Printf("CALL JobHistTab.View()\n")
		MainWindow.WriteString(m.JobHistTab.View(m.Log))

	case tabJobDetails:
		MainWindow.WriteString(m.tabJobDetails())

	case tabJobFromTemplate:
		MainWindow.WriteString(m.tabJobFromTemplate())

	case tabCluster:
		// Top Main
		MainWindow.WriteString(fmt.Sprintf("Filter: %10.20s\tItems: %d\n\n", m.JobClusterTab.Filter.Value(), len(m.JobClusterTab.SinfoFiltered.Nodes)))
		MainWindow.WriteString(m.tabClusterBars())

		// Mid Main: table || table+stats
		switch {
		case m.JobClusterTab.StatsOn:
			MainWindow.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.tabCluster(), styles.MenuBoxStyle.Render(m.JobClusterTabStats())))
		default:
			MainWindow.WriteString(m.tabCluster())
		}

		// Low Main: nil || filter || counts
		switch {
		case m.JobClusterTab.FilterOn:
			// filter
			MainWindow.WriteString("\n")
			// filter
			MainWindow.WriteString("\n")
			MainWindow.WriteString("Filter value (search across: Name, State, StateFlags!):\n")
			MainWindow.WriteString(fmt.Sprintf("%s\n", m.JobClusterTab.Filter.View()))
			MainWindow.WriteString("(Enter to apply, Esc to clear filter and abort, Regular expressions supported, syntax details: https://golang.org/s/re2syntax)\n")
		case m.JobClusterTab.CountsOn:
			MainWindow.WriteString("\n")
			MainWindow.WriteString(styles.JobInfoBox.Render(m.getClusterCounts()))

		default:
			MainWindow.WriteString("\n")
			MainWindow.WriteString(GenCountStr(m.JobClusterTab.Stats.StateCnt, m.Log))
		}

	case tabAbout:
		MainWindow.WriteString(m.tabAbout())
	}

	// FOOTER
	scr.WriteString(lipgloss.JoinVertical(lipgloss.Left, styles.MainWindow.Render(MainWindow.String()), styles.HelpWindow.Render(m.Help.View(keybindings.DefaultKeyMap))))

	return scr.String()
}
