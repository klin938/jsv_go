package main

import (
	"jsv"
)

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
	// So far this is the only way I can use to identify if a job is a batch (qsub) or
	// interactive (qrsh, qlogin etc). I assume most people use qrsh/qlogin without a
	// command (so to be brought to a remote compute node cmd promt). qrsh/qlogin can
	// still accept a command. In this case, the command runs remotely in real time and
	// the qrsh/qlogin session is terminated automatically, which no problem for us.
	cmdname, _ := jsv.GetParam("CMDNAME")
	if cmdname == "NONE" {
		jsv.LogInfo("JSV: " + client + "|" + job_id + "|" + cmdname + " - setting wall clock limits [ h_rt:8:00:00 | s_rt:8:00:00 ]")
		// existing param is overwritten automatically
		jsv.SubAddParam("l_hard", "h_rt", "8:00:00")
		jsv.SubAddParam("l_hard", "s_rt", "8:00:00")
	}

	tmp_req_val, _ := jsv.SubGetParam("l_hard", "tmp_requested")
	jsv.AddEnv("tmp_requested", tmp_req_val)

	mem_req_val, _ := jsv.SubGetParam("l_hard", "mem_requested")
	jsv.AddEnv("mem_requested", mem_req_val)
	// Basically we are setting a RAM quota here
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
