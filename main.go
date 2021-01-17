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
    cfg, err := ini.Load("shovel.ini")
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
        }

        container.Name = configIni.Section("config").Key("name").String()
        container.Description = configIni.Section("config").Key("description").String()
        container.Plugin = configIni.Section("config").Key("plugin").String()
        container.Argument = configIni.Section("config").Key("argument").String()
        interval, _ := time.ParseDuration(configIni.Section("config").Key("interval").String())
        container.Interval = int(interval.Seconds())
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
                go func(name, command, argument, notifier string){
                    log.Println("Running check: " + name)
                    cmd := exec.Command("/bin/sh", "-c", command + " " + argument)
                    var outputBuffer bytes.Buffer
                    cmd.Stdout = &outputBuffer
                    cmd.Env = os.Environ()
                    cmd.Env = append(cmd.Env,
                                        "PLUGINSDIR=" + pluginsDir)
                    if err := cmd.Run() ; err != nil {
                        if exitError, ok := err.(*exec.ExitError); ok {
                            configArray[index].CurrentStatus = exitError.ExitCode()
                        }
                    } else {
                        configArray[index].CurrentStatus = 0
                    }

                    if configArray[index].CurrentStatus != configArray[index].Status {
                        log.Println("Status of check " + 
                                    configArray[index].Name + 
                                    " changed from " + 
                                    configArray[index].Status +
                                    " to " +
                                    configArray[index].CurrentStatus)
                        alert := exec.Command("/bin/sh", "-c", notifier)
                        alert.Env = os.Environ()
                        alert.Env = append(alert.Env,
                                            "NAME=" + configArray[index].Name,
                                            "STATUS=" + strconv.Itoa(configArray[index].CurrentStatus),
                                            "DESCRIPTION=" + configArray[index].Description,
                                            "MESSAGE=" + outputBuffer.String())
                        err := alert.Run()
                        if err != nil {
                            log.Println("Unable to launch alert:" + err.Error())
                        }
                    }
                    configArray[index].Status = configArray[index].CurrentStatus
                }(configArray[index].Name,
                    pluginsDir + "/" + configArray[index].Plugin,
                    configArray[index].Argument,
                    notifiersDir + "/" + configArray[index].Notify)
                configArray[index].Counter = 0
            }
        }
        time.Sleep(1 * time.Second)
    }
}
