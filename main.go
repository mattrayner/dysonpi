package main

import (
	"context"
	"encoding/hex"
	"fmt"
	wemo "github.com/jdfergason/go.wemo"
	"log"
	"net"
	"strings"
	"time"
	"github.com/waringer/broadlink/broadlinkrm"
	"github.com/stianeikeland/go-rpio/v4"
)

func discover() {
	api, _ := wemo.NewByInterface("en0")
	devices, _ := api.DiscoverAll(3*time.Second)
	for _, device := range devices {
		fmt.Printf("Found %+v\n", device)
	}
}

var devices []broadlinkrm.Device
var switchOff = false

func triggerWemo(ctx context.Context) {
	device        := &wemo.Device{Host:"192.168.1.213:49153"}

	deviceInfo, _ := device.FetchDeviceInfo(ctx)
	log.Printf("Device Info => %+v\n", deviceInfo)

	state := device.GetBinaryState()
	log.Printf("Current State => %+v\n", state)

	if(state == 0) {
		log.Print("Switch was off, turning on.\n")
		device.On()

		log.Print("Sleeping for 3 seconds to allow switch to update")
		time.Sleep(3 * time.Second)

		log.Printf("Current State => %+v\n", device.GetBinaryState())

		switchOff = false

		triggerBroadlink(ctx)
	} else {
		if switchOff {
			log.Fatalln("Switch previously in off state, preventing endless loop by dying")
		}

		log.Print("Switch is on, toggling power.\n")
		device.Off()

		log.Print("Sleeping for 3 seconds to allow switch to update")
		time.Sleep(3 * time.Second)

		switchOff = true

		triggerWemo(ctx)
	}
}

func wemoOff(ctx context.Context) {
	log.Print("Turning off switch")
	device := &wemo.Device{Host:"192.168.1.213:49153"}
	device.Off()
}

func triggerBroadlink(ctx context.Context) {
	devices =  discoverBroadlink(net.ParseIP("192.168.1.84"))
	irCommand, err := hex.DecodeString(strings.Replace("26004800481917191818182e182e171917181818171a18191719171917191719172e181817000cc3481917191819172f172e181818181718181a17191719181818181818182d18 ", " ", "", -1))

	if err != nil {
		log.Fatalln("Provided Broadlink IR code is invalid")
	}

	if irCommand != nil {
		for id, device := range devices {
			id++

			response := broadlinkrm.Command(2, irCommand, &device)

			if response == nil {
				log.Print(response)
				log.Printf("[%02v] code send failed!\n", id)
			} else {
				log.Printf("[%02v] code sent.\n", id)
			}
		}
	}
}

func discoverBroadlink(ip net.IP) (dev []broadlinkrm.Device) {
	var devC chan broadlinkrm.Device

	devC = broadlinkrm.Hello(0, ip)

	id := 0
	for device := range devC {
		id++

		log.Print(fmt.Sprintf("[%02v] Device type: %X \n", id, device.DeviceType))
		log.Print(fmt.Sprintf("[%02v] Device name: %v \n", id, device.DeviceName))
		log.Print(fmt.Sprintf("[%02v] Device MAC: [% x] \n", id, device.DeviceMac()))
		log.Print(fmt.Sprintf("[%02v] Device IP: %v \n", id, device.DeviceAddr.IP))

		broadlinkrm.Auth(&device)

		log.Print(fmt.Sprintf("[%02v] Authenticated\n", id))

		dev = append(dev, device)
	}

	log.Print(fmt.Sprintf("Found %v device(s)\n", len(dev)))
	return
}

func main() {
	// retrieve device info
	ctx := context.Background()

	err := rpio.Open()
	if err != nil {
		log.Fatalln("Error opening GPIO pin")
	}

	defer rpio.Close()

	pin := rpio.Pin(23)
	pin.Input()
	pin.PullDown()

	loop := true
	var pin_high = pin.Read() == rpio.High
	log.Printf("Initial pin value: (high=%v)", pin_high)
	for loop {
		res := pin.Read()
		current_pin_high := res == rpio.High

		if current_pin_high != pin_high {
			pin_high = current_pin_high

			log.Printf("New pin value: (high=%v)", pin_high)

			if pin_high {
				triggerWemo(ctx)
			} else {
				triggerBroadlink(ctx)

				time.Sleep(3 * time.Second)

				wemoOff(ctx)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
