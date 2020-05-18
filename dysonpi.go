package main

import (
	"context"
	"fmt"
	wemo "github.com/jdfergason/go.wemo"
	"time"
)

func discover() {
	api, _ := wemo.NewByInterface("en0")
	devices, _ := api.DiscoverAll(3*time.Second)
	for _, device := range devices {
		fmt.Printf("Found %+v\n", device)
	}
}

func main() {
	device        := &wemo.Device{Host:"192.168.1.213:49153"}

	// retrieve device info
	ctx := context.Background()
	deviceInfo, _ := device.FetchDeviceInfo(ctx)
	fmt.Printf("Device Info => %+v\n", deviceInfo)

	state := device.GetBinaryState()
	fmt.Printf("Current State => %+v\n", state)

	deviceInfo, _ = device.FetchDeviceInfo(ctx)
	fmt.Printf("Device Info => %+v\n", deviceInfo)

	device.Toggle()

	state = device.GetBinaryState()
	fmt.Printf("Current State => %+v\n", state)

	deviceInfo, _ = device.FetchDeviceInfo(ctx)
	fmt.Printf("Device Info => %+v\n", deviceInfo)

	device.Toggle()

	state = device.GetBinaryState()
	fmt.Printf("Current State => %+v\n", state)

	deviceInfo, _ = device.FetchDeviceInfo(ctx)
	fmt.Printf("Device Info => %+v\n", deviceInfo)
}
