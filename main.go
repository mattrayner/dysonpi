package main

import (
	"context"
	"encoding/hex"
	"fmt"
	wemo "github.com/jdfergason/go.wemo"
	"log"
	"net"
	"time"
	"github.com/waringer/broadlink/broadlinkrm"
	"github.com/stianeikeland/go-rpio"
)

var wemoAddress = "192.168.1.207:49153"
var broadlinkAddress = "192.168.1.84"

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
	device        := &wemo.Device{Host:wemoAddress}

	deviceInfo, _ := device.FetchDeviceInfo(ctx)
	log.Printf("[WeMo][%s] Device IP: %v, connected.\n", deviceInfo.FriendlyName, device.Host)

	state := device.GetBinaryState()
	log.Printf("[WeMo][%s] State: %+s\n", deviceInfo.FriendlyName, prettyWemoState(state))

	if(state == 0) {
		log.Print("[WeMo] -> Switch was off, turning on.\n")
		device.On()

		log.Print("[WeMo] -> Sleeping for 3 seconds to allow switch to update.\n")
		time.Sleep(3 * time.Second)

		log.Printf("[WeMo][%s] State: %+s\n", deviceInfo.FriendlyName, prettyWemoState(device.GetBinaryState()))

		switchOff = false

		triggerBroadlink(ctx)
	} else {
		if switchOff {
			log.Fatalln("[WeMo] -> ! Switch previously OFF, exiting to prevent loop.")
		}

		log.Printf("[WeMo][%s] Turning off..\n", deviceInfo.FriendlyName)
		device.Off()

		log.Print("[WeMo] -> Sleeping for 3 seconds to allow switch to update.\n")
		time.Sleep(3 * time.Second)

		switchOff = true

		triggerWemo(ctx)
	}
}

func prettyWemoState(state int) string {
	var prettyState string
	if state == 0 {
		prettyState = "OFF"
	} else if state == 1 {
		prettyState = "ON"
	} else {
		log.Fatalln(fmt.Sprintf("Unexpected non-binary WeMo state received: %v", state))
	}

	return prettyState
}

func wemoOff(ctx context.Context) {
	device := &wemo.Device{Host:wemoAddress}

	deviceInfo, _ := device.FetchDeviceInfo(ctx)
	log.Printf("[WeMo][%s] Device IP: %v, connected.\n", deviceInfo.FriendlyName, device.Host)

	log.Printf("[WeMo][%s] Turning off..\n", deviceInfo.FriendlyName)
	device.Off()
	log.Print("[WeMo] -> Done")
}

func triggerBroadlink(ctx context.Context) {
	devices =  discoverBroadlink(net.ParseIP(broadlinkAddress))

	irCommand1, err1 := hex.DecodeString("26002400491b161b151b15311531151a161a151a161c151b161a161b151a161a161a153016000d0500000000")
	irCommand2, err2 := hex.DecodeString("26002400471a161a161b1530172f161a16191719151c163016301630153017191719161917000d0500000000")
	irCommand3, err3 := hex.DecodeString("26002400481917191819172e182e1818171818181730182e18181818182e182e1718181817000d0500000000")

	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatalln("[Broadlink] Provided Broadlink IR code is invalid")
	}

	codes := [3][]byte{irCommand1, irCommand2, irCommand3}
	for id, device := range devices {
		id++

		for codeId, code := range codes {
			codeId++
			time.Sleep(1500 * time.Millisecond)

			response := broadlinkrm.Command(2, code, &device)

			if response == nil {
				log.Print(response)
				log.Printf("[Broadlink][%02v-%02v] code send failed!\n", id, codeId)
			} else {
				log.Printf("[Broadlink][%02v-%02v] code sent.\n", id, codeId)
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

		log.Print(fmt.Sprintf("[Broadlink][%02v] Device IP: %v \n", id, device.DeviceAddr.IP))

		broadlinkrm.Auth(&device)

		log.Print(fmt.Sprintf("[Broadlink][%02v] Authenticated.\n", id))

		dev = append(dev, device)
	}

	if len(dev) == 0 {
		log.Fatalln("[Broadlink] No devices found.")
	}

	return
}

func setupLed(redPin int, greenPin int, bluePin int) {
	log.Printf("[LED] Red=%s Green=%s Blue=%s", redPin, greenPin, bluePin)

	red := rpio.Pin(redPin)
	//green := rpio.Pin(greenPin)
	//blue := rpio.Pin(bluePin)

	red.DutyCycle(1, 4)
	red.Freq(38000*4)
//	pin.DutyCycle(1, 4)
//	pin.Freq(38000*4)
//	pin.DutyCycle(1, 4)
//	pin.Freq(38000*4)
}

func main() {
	// retrieve device info
	ctx := context.Background()

	err := rpio.Open()
	if err != nil {
		log.Fatalln("[GPIO] Error opening GPIO pin")
	}

	defer rpio.Close()

	pin := rpio.Pin(25)
	log.Print("Sleeping so that pull down works on boot")
	time.Sleep(3 * time.Second)

	pin.Input()
	pin.PullDown()

	loop := true
	var pin_high = pin.Read() == rpio.High
	log.Printf("[GPIO] Initial pin value: (high=%v)", pin_high)

	//setupLed(18, 23, 24)
	//setLed(100, 100, 100)

	if pin_high {
		triggerWemo(ctx)
	}

	for loop {
		pin.PullDown()
		res := pin.Read()
		current_pin_high := res == rpio.High

		if current_pin_high != pin_high {
			pin_high = current_pin_high

			log.Printf("[GPIO] Pin value: (high=%v)", pin_high)

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
