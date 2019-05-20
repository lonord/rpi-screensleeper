package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/stianeikeland/go-rpio"
)

var (
	appName    = "screen-sleeper"
	appVersion = "dev"
	buildTime  = "unknow"
)

func main() {
	cfg := flag.String("config", "/etc/screen-sleeper/config.yml", "specify the config file")
	delay := flag.Int("delay", 0, "delay `seconds` for gpio input")
	pin := flag.Int("pin", 0, "bcm pin (not physical pin) number of `gpio`")
	trigger := flag.String("trigger", "", "trigger `edge` of gpio input, 'up' or 'down'")
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
	if *trigger != "" {
		c.Trigger = *trigger
	}
	err = run(c)
	if err != nil {
		printErrorAndExit(err)
	}
}

func run(c *config) error {
	// TODO delete it
	fmt.Println("**** using config ****")
	fmt.Println("delay:", c.Delay)
	fmt.Println("pin:", c.Pin)
	fmt.Println("trigger:", c.Trigger)
	err := rpio.Open()
	if err != nil {
		return err
	}
	defer rpio.Close()
	// TODO
	return nil
}

func printErrorAndExit(err error) {
	fmt.Println("error:", err.Error())
	os.Exit(1)
}
