package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	Delay   int    `yaml:"delay"`
	Pin     int    `yaml:"pin"`
	Trigger string `yaml:"trigger"`
}

func readConfig(cfg string) (*config, error) {
	b, err := ioutil.ReadFile(cfg)
	if err != nil {
		fmt.Println("warn: ignore non exist config file", cfg)
		return &config{}, nil
	}
	c := config{}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
