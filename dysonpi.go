package main

import (
	"context"
	"fmt"
	wemo "github.com/jdfergason/go.wemo"
	"time"
	"github.com/waringer/broadlink/broadlinkrm"
)

func discover() {
	api, _ := wemo.NewByInterface("en0")
	devices, _ := api.DiscoverAll(3*time.Second)
	for _, device := range devices {
		fmt.Printf("Found %+v\n", device)
	}
}

func triggerWemo(ctx context.Context) {
	device        := &wemo.Device{Host:"192.168.1.213:49153"}

	deviceInfo, _ := device.FetchDeviceInfo(ctx)
	fmt.Printf("Device Info => %+v\n", deviceInfo)

	state := device.GetBinaryState()
	fmt.Printf("Current State => %+v\n", state)

	fmt.Print("=========\n")

	if(state == 0) {
		fmt.Print("Switch was off, turning on.\n")
		device.On()

		fmt.Printf("Current State => %+v\n", device.GetBinaryState())
	} else {
		fmt.Print("Switch is on, no Wemo action needed.\n")
	}

	triggerBroadlink(ctx)
}

func triggerBroadlink(ctx context.Context) {
	fmt.Print("=========\n")
	fmt.Print("Fire BroadLink code.\n")
}

func main() {
	// retrieve device info
	ctx := context.Background()

	triggerWemo(ctx)
}
