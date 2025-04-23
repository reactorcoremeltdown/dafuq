package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gopkg.in/ini.v1"
)

type config struct {
	Name              string
	Plugin            string
	Argument          string
	Interval          int
	Description       string
	Notify            []string
	SuppressedBy      []string
	Output            string
	Counter           int
	DowntimeCounter   int
	Status            int
	CurrentStatus     int
	WarningThreshold  string
	CriticalThreshold string
	FlowOperator      string
	Hostname          string
}

var configArray []config
var Version, CommitID, BuildDate string
var debug bool

func logErr(desc string, e error) {
	if e != nil {
		log.Println(desc + ": " + e.Error())
	}
}

func displayVersion(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Dafuq, version: "+
		Version+
		", build date: "+BuildDate+
		", commit ID: "+CommitID+"\n")
}

func getCheck(name string) (config, error) {
	for _, check := range configArray {
		if check.Name == name {
			return check, nil
		}
	}
	return config{}, fmt.Errorf("No check with name %s found\n", name)
}

func encodeConfig(res http.ResponseWriter, req *http.Request) {
	checkName := req.URL.Query().Get("check")
	counter := req.URL.Query().Get("counter")
	downtime := req.URL.Query().Get("downtime")
	notFound := true
	if debug {
		fmt.Pritnf("[DEBUG] check: %s, counter: %s, downtime: %s\n", checkName, counter, downtime)
	}
	if checkName != "" {
		if req.Method == http.MethodPost {
			if counter != "" {
				counterValue, err := strconv.Atoi(counter)
				if err != nil {
					res.WriteHeader(400)
					fmt.Fprint(res, "Invalid integer value\n")
				} else {
					for index, check := range configArray {
						if check.Name == checkName {
							notFound = false
							configArray[index].Counter = counterValue
							fmt.Fprint(res, "OK\n")
						}
					}
					if notFound {
						res.WriteHeader(404)
						fmt.Fprint(res, "Check not found\n")
					}
				}
			} else if downtime != "" {
				downtimeValue, err := time.ParseDuration(downtime)
				if err != nil {
					res.WriteHeader(400)
					fmt.Fprint(res, "Failed to parse downtime duration\n")
				} else {
					downtimeCounter := int(downtimeValue.Seconds())
					for index, check := range configArray {
						if check.Name == checkName {
							notFound = false
							configArray[index].DowntimeCounter = downtimeCounter
							fmt.Fprint(res, "OK\n")
						}
					}
					if notFound {
						res.WriteHeader(404)
						fmt.Fprint(res, "Check not found\n")
					}
				}
			} else {
				res.WriteHeader(400)
				fmt.Fprint(res, "POST requests are only for setting check counters and downtime intervals\n")
			}
		} else {
			/*
				for _, check := range configArray {
					if check.Name == checkName {
						notFound = false
						configJson, err := json.Marshal(check)
						logErr("Cannot encode to JSON", err)
						fmt.Fprint(res, string(configJson))
					}
				}
				if notFound {
					res.WriteHeader(404)
					fmt.Fprint(res, "Check not found\n")
				}
			*/
			check, err := getCheck(checkName)
			if err != nil {
				res.WriteHeader(404)
				fmt.Fprint(res, "Check not found\n")
			} else {
				configJson, err := json.Marshal(check)
				logErr("Cannot encode to JSON", err)
				fmt.Fprint(res, string(configJson))
			}
		}
	} else {
		configJson, err := json.Marshal(configArray)
		logErr("Cannot encode to JSON", err)
		fmt.Fprint(res, string(configJson))
	}
}

func writeStateFile(path string) error {
	configJson, err := json.Marshal(configArray)
	if err != nil {
		return err
	}
	data := []byte(configJson)
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadState(loaded []config) {
	for index, _ := range configArray {
		for _, key_loaded := range loaded {
			if configArray[index].Name == key_loaded.Name {
				configArray[index].Counter = key_loaded.Counter
				configArray[index].DowntimeCounter = key_loaded.DowntimeCounter
				configArray[index].Status = key_loaded.Status
				configArray[index].CurrentStatus = key_loaded.CurrentStatus
				configArray[index].Output = key_loaded.Output
			}
		}
	}

	log.Println("Loading state completed")
}

func main() {
	log.Println(os.Environ())
	var configPath string
	configPathFromEnv, configPathFromEnvPresent := os.LookupEnv("CONFIG_PATH")
	if configPathFromEnvPresent {
		configPath = configPathFromEnv
	} else {
		configPath = "/etc/dafuq/config.ini"
	}
	cfg, err := ini.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load config file: %v", err)
		os.Exit(1)
	}

	configsDir := cfg.Section("main").Key("configs").String()
	pluginsDir := cfg.Section("main").Key("plugins").String()
	notifiersDir := cfg.Section("main").Key("notifiers").String()
	stateFilePath := cfg.Section("main").Key("stateFile").String()
	execTimeoutSec := cfg.Section("main").Key("execTimeoutSec").MustInt(10) // Defaulting to 10 seconds timeout for executing scripts
	jsonStatusPath := cfg.Section("main").Key("jsonStatusPath").MustString("/")
	address := cfg.Section("main").Key("address").String()
	debug = cfg.Section("main").Key("debug").MustBool(false)
	port := cfg.Section("main").Key("port").String()

	configFiles, err := ioutil.ReadDir(configsDir + "/")
	if err != nil {
		fmt.Printf("Failed to read directory contents: %v", err)
		os.Exit(1)
	}

	for _, configFile := range configFiles {
		var container config
		configIni, err := ini.ShadowLoad(configsDir + "/" + configFile.Name())
		if err != nil {
			log.Println("Failed to parse config file: " + err.Error())
		} else {
			log.Println("Loaded config file: " + configsDir + "/" + configFile.Name())
		}

		container.Name = configIni.Section("config").Key("name").String()
		container.Description = configIni.Section("config").Key("description").String()
		container.Plugin = configIni.Section("config").Key("plugin").String()
		container.Argument = configIni.Section("config").Key("argument").String()
		container.Hostname = configIni.Section("config").Key("hostname").MustString(os.Getenv("HOSTNAME"))
		interval, _ := time.ParseDuration(configIni.Section("config").Key("interval").String())
		seconds := int(interval.Seconds())
		if seconds < 5 {
			container.Interval = 5
		} else {
			container.Interval = seconds
		}
		if configIni.Section("config").Key("warningThreshold").String() != "" {
			container.WarningThreshold = configIni.Section("config").Key("warningThreshold").String()
		} else {
			container.WarningThreshold = "0"
		}
		if configIni.Section("config").Key("criticalThreshold").String() != "" {
			container.CriticalThreshold = configIni.Section("config").Key("criticalThreshold").String()
		} else {
			container.CriticalThreshold = "0"
		}
		if configIni.Section("config").Key("flowOperator").String() != "" {
			container.FlowOperator = configIni.Section("config").Key("flowOperator").String()
		} else {
			container.FlowOperator = "upwards"
		}
		container.Notify = configIni.Section("config").Key("notify").ValueWithShadows()
		container.SuppressedBy = configIni.Section("config").Key("suppressedBy").ValueWithShadows()
		container.Output = "Waiting for output"
		container.Counter = 0
		container.DowntimeCounter = 0
		container.Status = 0
		container.CurrentStatus = 0

		configArray = append(configArray, container)
		container.WarningThreshold = "0"
		container.CriticalThreshold = "0"
		container.FlowOperator = "upwards"
	}

	loadedState := make([]config, 0)
	stateData, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		log.Println("Unable to load state from file: " + err.Error())
	} else {
		err = json.Unmarshal(stateData, &loadedState)
		if err != nil {
			log.Println("Unable to decode JSON from state data: " + err.Error())
		} else {
			loadState(loadedState)
			log.Println("Loaded state from " + stateFilePath)
		}
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		err := writeStateFile(stateFilePath)
		logErr("Unable to write data to state file", err)
		os.Exit(0)
	}()

	go func() {
		http.HandleFunc("/version", displayVersion)
		http.HandleFunc(jsonStatusPath, encodeConfig)
		log.Println(http.ListenAndServe(address+":"+port, nil))
	}()
	for {
		for index, _ := range configArray {
			configArray[index].Counter = configArray[index].Counter + 1
			configArray[index].DowntimeCounter = configArray[index].DowntimeCounter - 1
			if configArray[index].DowntimeCounter < 0 {
				configArray[index].DowntimeCounter = 0
			}
			if configArray[index].Counter >= configArray[index].Interval {
				go func(i int) {
					log.Println("Running check: " + configArray[i].Name)
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(execTimeoutSec)*time.Second)
					cmd := exec.CommandContext(ctx, "/bin/sh", "-c", pluginsDir+"/"+configArray[i].Plugin+" "+configArray[i].Argument)
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					go func() {
						<-ctx.Done()
						if ctx.Err() == context.DeadlineExceeded {
							syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
						}
					}()
					var outputBuffer, stderrBuffer bytes.Buffer
					cmd.Stdout = &outputBuffer
					cmd.Stderr = &stderrBuffer
					cmd.Env = os.Environ()
					cmd.Env = append(cmd.Env,
						"WARNING_THRESHOLD="+configArray[i].WarningThreshold,
						"CRITICAL_THRESHOLD="+configArray[i].CriticalThreshold,
						"FLOW_OPERATOR="+configArray[i].FlowOperator,
						"PLUGIN_NAME="+configArray[i].Name,
						"PLUGINSDIR="+pluginsDir,
						"HOSTNAME="+configArray[i].Hostname)
					if err := cmd.Run(); err != nil {
						if exitError, ok := err.(*exec.ExitError); ok {
							configArray[i].CurrentStatus = exitError.ExitCode()
						}
					} else {
						configArray[i].CurrentStatus = 0
					}
					cancel()

					if ctx.Err() == context.DeadlineExceeded {
						log.Println("Timeout running check " + pluginsDir + "/" + configArray[i].Plugin)
					}
					configArray[i].Output = outputBuffer.String()

					if stderrBuffer.String() != "" {
						log.Println("Check " + configArray[i].Name + " errored: " + stderrBuffer.String())
					}

					if configArray[i].CurrentStatus != configArray[i].Status {
						err := writeStateFile(stateFilePath)
						logErr("Unable to write data to state file", err)
						log.Println("Status of check " +
							configArray[i].Name +
							" changed from " +
							strconv.Itoa(configArray[i].Status) +
							" to " +
							strconv.Itoa(configArray[i].CurrentStatus))
						suppressStatus := 0
						for _, suppressor := range configArray[i].SuppressedBy {
							s, err := getCheck(suppressor)
							if err == nil {
								suppressStatus = suppressStatus + s.Status
							}
						}

						if suppressStatus == 0 || configArray[i].DowntimeCounter != 0 {
							for _, item := range configArray[i].Notify {
								ctx, cancel := context.WithTimeout(context.Background(), time.Duration(execTimeoutSec)*time.Second)
								alert := exec.CommandContext(ctx, "/bin/sh", "-c", notifiersDir+"/"+item)
								alert.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
								go func() {
									<-ctx.Done()
									if ctx.Err() == context.DeadlineExceeded {
										syscall.Kill(-alert.Process.Pid, syscall.SIGKILL)
									}
								}()
								alert.Env = os.Environ()
								alert.Env = append(alert.Env,
									"NAME="+configArray[i].Name,
									"STATUS="+strconv.Itoa(configArray[i].CurrentStatus),
									"HOSTNAME="+configArray[i].Hostname,
									"DESCRIPTION="+configArray[i].Description,
									"MESSAGE="+outputBuffer.String())
								err = alert.Run()
								cancel()
								if err != nil {
									log.Println("Command is: " + notifiersDir + "/" + item)
									log.Println("Unable to launch alert", err)
								}
								if ctx.Err() == context.DeadlineExceeded {
									log.Println("Timeout running " + notifiersDir + "/" + item)
								}
							}
						} else {
							log.Println("Check " + configArray[i].Name + " suppressed")
						}
					}
					configArray[i].Status = configArray[i].CurrentStatus
				}(index)
				configArray[index].Counter = 0
			}
		}
		time.Sleep(1 * time.Second)
	}
}
