package main

import (
    "fmt"
    "time"
    "os"
    "os/exec"
    "io/ioutil"
    "strconv"
    "bytes"
    "log"

    "gopkg.in/ini.v1"
)

type config struct {
    Name string
    Plugin string
    Argument string
    Interval int
    Description string
    Notify string
    Counter int
    Status int
    CurrentStatus int
}

var configArray []config

func main() {
    cfg, err := ini.Load("/etc/monitoring/config.ini")
    if err != nil {
        fmt.Printf("Failed to load config file: %v", err)
        os.Exit(1)
    }

    configsDir := cfg.Section("main").Key("configs").String()
    pluginsDir := cfg.Section("main").Key("plugins").String()
    notifiersDir := cfg.Section("main").Key("notifiers").String()

    configFiles, err := ioutil.ReadDir(configsDir + "/")
    if err != nil {
        fmt.Printf("Failed to read directory contents: %v", err)
        os.Exit(1)
    }

    for _, configFile := range configFiles {
        var container config
        configIni, err := ini.Load(configsDir + "/" + configFile.Name())
        if err != nil {
            log.Println("Failed to parse config file: " + err.Error())
        } else {
            log.Println("Loaded config file: " + configsDir + "/" + configFile.Name())
        }

        container.Name = configIni.Section("config").Key("name").String()
        container.Description = configIni.Section("config").Key("description").String()
        container.Plugin = configIni.Section("config").Key("plugin").String()
        container.Argument = configIni.Section("config").Key("argument").String()
        interval, _ := time.ParseDuration(configIni.Section("config").Key("interval").String())
        seconds := int(interval.Seconds())
        if seconds < 5 {
            container.Interval = 5
        } else {
            container.Interval = seconds
        }
        container.Notify = configIni.Section("config").Key("notify").String()
        container.Counter = 0
        container.Status = 0
        container.CurrentStatus = 0

        configArray = append(configArray, container)
    }
    for {
        for index, _ := range configArray {
            configArray[index].Counter = configArray[index].Counter + 1
            if (configArray[index].Counter == configArray[index].Interval) {
                go func(i int){
                    log.Println("Running check: " + configArray[i].Name)
                    cmd := exec.Command("/bin/sh", "-c", pluginsDir + "/" + configArray[i].Plugin + " " + configArray[i].Argument)
                    var outputBuffer bytes.Buffer
                    cmd.Stdout = &outputBuffer
                    cmd.Env = os.Environ()
                    cmd.Env = append(cmd.Env,
                                        "PLUGINSDIR=" + pluginsDir)
                    if err := cmd.Run() ; err != nil {
                        if exitError, ok := err.(*exec.ExitError); ok {
                            configArray[i].CurrentStatus = exitError.ExitCode()
                        }
                    } else {
                        configArray[i].CurrentStatus = 0
                    }

                    if configArray[i].CurrentStatus != configArray[i].Status {
                        log.Println("Status of check " + 
                                    configArray[i].Name + 
                                    " changed from " + 
                                    strconv.Itoa(configArray[i].Status) +
                                    " to " +
                                    strconv.Itoa(configArray[i].CurrentStatus))
                        alert := exec.Command("/bin/sh", "-c", notifiersDir + "/" + configArray[i].Notify)
                        alert.Env = os.Environ()
                        alert.Env = append(alert.Env,
                                            "NAME=" + configArray[i].Name,
                                            "STATUS=" + strconv.Itoa(configArray[i].CurrentStatus),
                                            "DESCRIPTION=" + configArray[i].Description,
                                            "MESSAGE=" + outputBuffer.String())
                        err := alert.Run()
                        if err != nil {
                            log.Println("Unable to launch alert:" + err.Error())
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
