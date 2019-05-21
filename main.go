package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio"
)

var (
	appName    = "screen-sleeper"
	appVersion = "dev"
	buildTime  = "unknow"
)

const screenDev = "/sys/class/backlight/rpi_backlight/bl_power"

func main() {
	cfg := flag.String("config", "/etc/screen-sleeper/config.yml", "specify the config file")
	delay := flag.Int("delay", 0, "delay `seconds` for gpio input")
	pin := flag.Int("pin", 0, "bcm pin (not physical pin) number of `gpio`")
	level := flag.String("level", "", "enable `level` of gpio input, 'high' or 'low'")
	ver := flag.Bool("version", false, "show version")
	flag.Parse()
	if *ver {
		fmt.Println("version", appVersion)
		fmt.Println("build time", buildTime)
		os.Exit(1)
	}
	c, err := readConfig(*cfg)
	if err != nil {
		printErrorAndExit(err)
	}
	if *delay > 0 {
		c.Delay = *delay
	}
	if *pin > 0 {
		c.Pin = *pin
	}
	if *level != "" {
		c.Level = *level
	}
	if c.Level == "" {
		c.Level = "high"
	}
	if c.Level != "high" && c.Level != "low" {
		printErrorAndExit(errors.New("invalid level value, 'high' or 'low' is required"))
	}
	err = run(c)
	if err != nil {
		printErrorAndExit(err)
	}
}

func run(c *config) error {
	err := writeScreen(true)
	if err != nil {
		return err
	}
	defer writeScreen(true)
	err = rpio.Open()
	if err != nil {
		return err
	}
	defer rpio.Close()
	// setup pin
	pin := rpio.Pin(c.Pin)
	pin.Input()
	if c.Level == "high" {
		pin.PullDown()
	} else {
		pin.PullUp()
	}
	pin.Detect(rpio.AnyEdge)
	stat := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)
	go startCheck(ctx, stat, pin, c.Level, c.Delay)
	for {
		select {
		case s := <-stat:
			if err := writeScreen(s); err != nil {
				return err
			}
		case sig := <-signalChan:
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				return nil
			}
		}
	}
}

func startCheck(ctx context.Context, stat chan bool, pin rpio.Pin, level string, delay int) {
	timer := time.NewTimer(time.Second * time.Duration(delay))
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if !readPinStat(pin, level) {
				stat <- false
			}
		default:
			if pin.EdgeDetected() {
				if readPinStat(pin, level) {
					stat <- true
				} else {
					timer.Reset(time.Second * time.Duration(delay))
				}
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func readPinStat(pin rpio.Pin, level string) bool {
	return level == "high" && pin.Read() == rpio.High || level == "low" && pin.Read() == rpio.Low
}

func writeScreen(on bool) error {
	if _, err := os.Stat(screenDev); err != nil {
		return err
	}
	var data string
	if on {
		data = "0"
	} else {
		data = "1"
	}
	return ioutil.WriteFile(screenDev, []byte(data), 0644)
}

func printErrorAndExit(err error) {
	fmt.Println("error:", err.Error())
	os.Exit(1)
}
