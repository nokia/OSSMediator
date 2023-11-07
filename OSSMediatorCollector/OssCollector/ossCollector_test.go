/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package OssCollector

/*func TestInitLogger(t *testing.T) {
	logDir = "./log"
	logLevel = 5
	initLogger(logDir, logLevel)
	logFile := logDir + "/collector.log"
	defer os.RemoveAll(logDir)
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fail()
	}
}
func TestParseFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{
		"cmd",
		"-log_dir", "./log",
		"-conf_file", "conf.json",
		"-cert_file", "tmp.crt",
		"-log_level", "3",
		"-skip_tls", "true",
		"-enable_console_log", "true",
	}
	parseFlags()

	if logDir != "./log" || confFile != "conf.json" || certFile != "tmp.crt" || logLevel != 3 {
		t.Fail()
	}
}
*/
