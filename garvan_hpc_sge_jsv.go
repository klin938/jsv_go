package main

import (
	"fmt"
	"jsv"
	"strings"
	"strconv"
	"time"
)

// 24:00:00/86400
// 8:00:00/28800
var DEFAULT_RT string = "8:00:00"
var DEFAULT_RT_SECONDS int = 28800
var SHORT_Q_RT string = "48:00:00"
var SHORT_Q_RT_SECONDS int = 172800

func jsvOnStartFunction() {
	//jsv_send_env()
}

func jsvVerificationFunction() {
	// /tmp/jsv_logfile.log
	jsv.LoggingEnabled = true
	// This is pretty confusing param. It actually means who is calling me(jsv). So
	// it is useful only in client side jsv to tell qsub, qrsh, qlogin etc. When is
	// called by server side, the value is 'qmaster' regardless qsub or qrsh or qlogin.
	client, _ := jsv.GetParam("CLIENT")
	// This param is available only in server side jsv. Empty str when in client side.
	job_id, _ := jsv.GetParam("JOB_ID")

	h_rt_val, h_rt_set := jsv.SubGetParam("l_hard", "h_rt")
	if h_rt_set == false {
		jsv.LogInfo("JSV: " + client + "|" + job_id + "|h_rt - set default: " + SHORT_Q_RT)
		h_rt_val = DEFAULT_RT
	} else {
		var rt_seconds int = 0
		// In server side JSV, any h_rt being processed by the SGE qmaster has been
		// converted to second. In order to make it usable by both server and client
		// side, try converting to int first, if it failed, then do the time parsing.
		rt_seconds, err := strconv.Atoi(h_rt_val)
		if err != nil {
			// convert hh:mm:ss format into seconds
			rt_spl := strings.Split(h_rt_val, ":")
			rt_str := rt_spl[0] + "h" + rt_spl[1] + "m" + rt_spl[2] + "s"
			rt, _ := time.ParseDuration(rt_str)
			rt_seconds = int(rt.Seconds())
		}
		jsv.LogInfo("JSV: " + client + "|" + job_id + "|h_rt - user requested " + h_rt_val + "/" + fmt.Sprint(rt_seconds))

		if rt_seconds > SHORT_Q_RT_SECONDS {
			jsv.LogInfo("JSV: " + client + "|" + job_id + "|h_rt - modified: " + h_rt_val + " -> " + SHORT_Q_RT)
			h_rt_val = SHORT_Q_RT
		}
	}
	// If the user did not provide s_rt, set it to whatever h_rt is.
	s_rt_val, s_rt_set := jsv.SubGetParam("l_hard", "s_rt")
	if s_rt_set == false {
		s_rt_val = h_rt_val
	}
	// So far this is the only way I can use to identify if a job is a batch (qsub) or
	// interactive (qrsh, qlogin etc). I assume most people use qrsh/qlogin without a
	// command (so to be brought to a remote compute node cmd promt). qrsh/qlogin can
	// still accept a command. In this case, the command runs remotely in real time and
	// the qrsh/qlogin session is terminated automatically, which no problem for us.
	cmdname, _ := jsv.GetParam("CMDNAME")
	if cmdname == "NONE" {
		jsv.SubAddParam("l_hard", "h_rt", h_rt_val)
		jsv.SubAddParam("l_hard", "s_rt", s_rt_val)
		jsv.LogInfo("JSV: " + client + "|" + job_id + "|" + cmdname + " - setting wall clock limits [ h_rt:" + h_rt_val + " | s_rt:" + s_rt_val + " ]")
	}

	tmp_req_val, _ := jsv.SubGetParam("l_hard", "tmp_requested")
	jsv.AddEnv("tmp_requested", tmp_req_val)

	mem_req_val, _ := jsv.SubGetParam("l_hard", "mem_requested")
	jsv.AddEnv("mem_requested", mem_req_val)
	// make sure h_vmem is the same as mem_requested
	jsv.SubAddParam("l_hard", "h_vmem", mem_req_val)
	jsv.LogInfo("JSV: " + client + "|" + job_id + "|" + cmdname + " - setting memory hard limit [ h_vmem:" + mem_req_val + " ]")
	// accepting the job but indicating that we did
	// some changes
	jsv.Correct("Job was modified")
	return
}

func main() {
	jsv.Run(true, jsvVerificationFunction, jsvOnStartFunction)
}
