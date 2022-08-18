//go:generate ../../../tools/readme_config_includer/generator
package asterisk

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"strings"
	_ "embed"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Asterisk struct {
	Socket string `toml:"socket"`
	Nodeid string `toml:"nodeid"`
}

//go:embed sample.conf
var sampleConfig string

func (s *Asterisk) SampleConfig() string {
	return sampleConfig
}

func (s *Asterisk) Description() string {
	return "Collects metrics from Asterisk Open Source VoIP Software"
}

/*
func readFromSocket(sock net.Conn, c chan string) {
	var tmp = make([]byte, 256)
	var recvData = make([]byte, 2048)
	for {
		numBytes, err := sock.Read(tmp)
		if err != nil {
			break
		}
		recvData = append(recvData, tmp[:numBytes]...)
	}
	c <- string(bytes.Trim(recvData[:len(recvData)], "\x00"))

}
*/

func readFromSocketIO(sock net.Conn, c chan string) {
	response, _ := ioutil.ReadAll(sock)
	c <- string(bytes.Trim(response, "\x00"))
}

func asteriskCommand(message string, socket string) (string, error) {
	var buffChan = make(chan string)
	var err error
	message = "cli quit after " + message + "\000"
	buff := make([]byte, 2048)
	conn, err := net.Dial("unix", socket)

	if err != nil {
		panic(err)
	}
	_, _ = conn.Read(buff)
	conn.Write([]byte(message))
	go readFromSocketIO(conn, buffChan)
	var data = <-buffChan
	conn.Close()
	if strings.Contains(data, "No such command") {
		err = errors.New("No such command")
	}
	return data, err

}

func processCoreShowCalls(message string) (int64, int64) {
	activeCalls := int64(-1)
	totalCalls := int64(-1)
	messageSplit := strings.Split(strings.Replace(message, "\r\n", "\n", -1), "\n")
	re := regexp.MustCompile(``)
	for _, messageItem := range messageSplit {
		if strings.Contains(messageItem, "capacity") {
			re = regexp.MustCompile(`([0-9]+) of ([0-9]+) max active call[s]?`)
			activeCalls, _ = strconv.ParseInt(strings.TrimSpace(re.FindAllStringSubmatch(message, -1)[0][1]), 10, 64)
		}
		if strings.Contains(messageItem, "processed") {
			re = regexp.MustCompile(`([0-9]+) call[s]? processed`)
			totalCalls, _ = strconv.ParseInt(strings.TrimSpace(re.FindAllStringSubmatch(message, -1)[0][1]), 10, 64)
		}
	}
	return activeCalls, totalCalls
}

func processPeers(peers string) (int64, int64, int64, int64, int64) {
	totalPeers := int64(0)
	monitoredOnline := int64(0)
	monitoredOffline := int64(0)
	unmonitoredOnline := int64(0)
	unmonitoredOffline := int64(0)
	peersSplit := strings.Split(strings.Replace(peers, "\r\n", "\n", -1), "\n")
	re := regexp.MustCompile(``)

	for _, peersItem := range peersSplit {
		if strings.Contains(peersItem, "sip peers [Monitored:") {

			//get total peers
			re = regexp.MustCompile(`([0-9]+) sip peer`)
			totalPeers, _ = strconv.ParseInt(strings.TrimSpace(re.FindAllStringSubmatch(peersItem, -1)[0][1]), 10, 64)

			//get Monitored
			re = regexp.MustCompile(`Monitored: ([0-9]+) online, ([0-9]+) offline`)
			monitored := re.FindAllStringSubmatch(peersItem, -1)
			monitoredOnline, _ = strconv.ParseInt(strings.TrimSpace(monitored[0][1]), 10, 64)
			monitoredOffline, _ = strconv.ParseInt(strings.TrimSpace(monitored[0][2]), 10, 64)

			//get Unmonitored
			re = regexp.MustCompile(`Monitored: ([0-9]+) online, ([0-9]+) offline`)
			unmonitored := re.FindAllStringSubmatch(peersItem, -1)
			unmonitoredOnline, _ = strconv.ParseInt(strings.TrimSpace(unmonitored[0][1]), 10, 64)
			unmonitoredOffline, _ = strconv.ParseInt(strings.TrimSpace(unmonitored[0][2]), 10, 64)
		}
	}
	return totalPeers, monitoredOnline, monitoredOffline, unmonitoredOnline, unmonitoredOffline

}

func processUptime(uptime string) (int64, int64) {
	systemUptime := int64(0)
	systemUptimeYears := int64(0)
	systemUptimeWeeks := int64(0)
	systemUptimeDays := int64(0)
	systemUptimeHours := int64(0)
	systemUptimeMinutes := int64(0)
	systemUptimeSeconds := int64(0)

	asteriskLastReload := int64(0)
	asteriskLastReloadYears := int64(0)
	asteriskLastReloadWeeks := int64(0)
	asteriskLastReloadDays := int64(0)
	asteriskLastReloadHours := int64(0)
	asteriskLastReloadMinutes := int64(0)
	asteriskLastReloadSeconds := int64(0)

	uptimeSplit := strings.Split(strings.Replace(uptime, "\r\n", "\n", -1), "\n")
	re := regexp.MustCompile(``)

	for _, uptimeItem := range uptimeSplit {
		if strings.Contains(uptimeItem, "System uptime") {

			re = regexp.MustCompile(`([0-9]+) year`)
			systemYears := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemYears) > 0 {
				systemUptimeYears, _ = strconv.ParseInt(systemYears[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) week`)
			systemWeeks := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemWeeks) > 0 {
				systemUptimeWeeks, _ = strconv.ParseInt(systemWeeks[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) day`)
			systemDays := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemDays) > 0 {
				systemUptimeDays, _ = strconv.ParseInt(systemDays[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) hour`)
			systemHours := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemHours) > 0 {
				systemUptimeHours, _ = strconv.ParseInt(systemHours[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) minute`)
			systemMinutes := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemMinutes) > 0 {
				systemUptimeMinutes, _ = strconv.ParseInt(systemMinutes[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) second`)
			systemSeconds := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(systemSeconds) > 0 {
				systemUptimeSeconds, _ = strconv.ParseInt(systemSeconds[0][1], 10, 64)
			}
		}
		if strings.Contains(uptimeItem, "Last reload") {

			re = regexp.MustCompile(`([0-9]+) year`)
			reloadYears := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadYears) > 0 {
				asteriskLastReloadYears, _ = strconv.ParseInt(reloadYears[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) week`)
			reloadWeeks := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadWeeks) > 0 {
				asteriskLastReloadWeeks, _ = strconv.ParseInt(reloadWeeks[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) day`)
			reloadDays := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadDays) > 0 {
				asteriskLastReloadDays, _ = strconv.ParseInt(reloadDays[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) hour`)
			reloadHours := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadHours) > 0 {
				asteriskLastReloadHours, _ = strconv.ParseInt(reloadHours[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) minute`)
			reloadMinutes := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadMinutes) > 0 {
				asteriskLastReloadMinutes, _ = strconv.ParseInt(reloadMinutes[0][1], 10, 64)
			}

			re = regexp.MustCompile(`([0-9]+) second`)
			reloadSeconds := re.FindAllStringSubmatch(uptimeItem, -1)
			if len(reloadSeconds) > 0 {
				asteriskLastReloadSeconds, _ = strconv.ParseInt(reloadSeconds[0][1], 10, 64)
			}
		}
	}

	systemUptime = (systemUptimeYears * 31104000) + (systemUptimeWeeks * 604800) + (systemUptimeDays * 86400) + (systemUptimeHours * 3600) + (systemUptimeMinutes * 60) + systemUptimeSeconds

	asteriskLastReload = (asteriskLastReloadYears * 31104000) + (asteriskLastReloadWeeks * 604800) + (asteriskLastReloadDays * 86400) + (asteriskLastReloadHours * 3600) + (asteriskLastReloadMinutes * 60) + asteriskLastReloadSeconds

	return systemUptime, asteriskLastReload
}

func (ast *Asterisk) Gather(acc telegraf.Accumulator) error {

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	data, err := asteriskCommand("core show calls", ast.Socket)
	if err == nil {
		activeCalls, totalCalls := processCoreShowCalls(data)
		fields["active_calls"] = activeCalls
		fields["total_calls"] = totalCalls
	}
	peers, err := asteriskCommand("sip show peers", ast.Socket)
	if err == nil {
		totalPeers, monitoredOnline, monitoredOffline, unmonitoredOnline, unmonitoredOffline := processPeers(peers)
		fields["sip_peers"] = totalPeers
		fields["sip_monitored_online"] = monitoredOnline
		fields["sip_monitored_offline"] = monitoredOffline
		fields["sip_unmonitored_online"] = unmonitoredOnline
		fields["sip_unmonitored_offline"] = unmonitoredOffline
	}

	uptime, err := asteriskCommand("core show uptime", ast.Socket)
	if err == nil {
		system_uptime, last_reload := processUptime(uptime)
		fields["system_uptime"] = system_uptime
		fields["last_reload"] = last_reload
	}

	if ast.Nodeid != "" {
		tags["nodeid"] = ast.Nodeid
	}
	acc.AddFields("asterisk", fields, tags)
	return nil
}

func init() {
	inputs.Add("asterisk", func() telegraf.Input { return &Asterisk{Socket: "/var/run/asterisk/asterisk.ctl", Nodeid: ""} })
}
