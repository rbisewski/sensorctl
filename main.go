package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var (
	printVersion = false
	version      = "0.0"

	debugMode = false

	cpuinfoDirectory         = "/proc/cpuinfo"
	hardwareMonitorDirectory = "/sys/class/hwmon/"

	prefixes    = []string{"temp", "fan"}
	inputSuffix = "_input"

	hardwareNameFile = "name"

	// flag to check whether the AMD digital thermo module is in use
	digitalAmdPowerModuleInUse = false
)

func init() {
	flag.BoolVar(&printVersion, "version", false,
		"Print the current version of this program and exit.")

	flag.BoolVar(&debugMode, "debug", false,
		"Dump debug output to stdout.")
}

func main() {

	flag.Parse()

	if printVersion {
		fmt.Println("sensorctl v" + version)
		os.Exit(0)
	}

	// normally there will likely be at least one sensor exposed to
	// the operating system; however, in theory there could be edge cases
	// where there are no sensors, so account for that here
	listOfDeviceDirs, err := ioutil.ReadDir(hardwareMonitorDirectory)
	if err != nil {
		panic(err)
	}

	if debugMode {

		debug("The following IDs are present in the hardware sensor " +
			"monitoring directory:\n")

		for _, dir := range listOfDeviceDirs {
			debug("* " + dir.Name())
		}
	}

	// Search thru the directories and set the relevant flags...
	err = SetGlobalSensorFlags(listOfDeviceDirs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// For each of the devices...
	for _, dir := range listOfDeviceDirs {

		// Assemble the filepath to the name file of the currently given
		// hardware device.
		hardwareNameFilepathOfGivenDevice := hardwareMonitorDirectory +
			dir.Name() + "/" + hardwareNameFile

		// If debug mode, print out the current 'name' file we are about
		// to open.
		debug(dir.Name() + " --> " + hardwareNameFilepathOfGivenDevice)

		// check to see if a 'name' file is present inside the directory.
		nameValueOfHardwareDevice, err := ioutil.ReadFile(hardwareNameFilepathOfGivenDevice)
		if err != nil {
			debug("Warning: " + dir.Name() + " does not contain a " +
				"hardware name file. Skipping...")
			continue
		}

		if len(nameValueOfHardwareDevice) < 1 {
			debug("Warning: The hardware name file of " + dir.Name() +
				" does not contain valid data. Skipping...")
			continue
		}

		// Trim away any excess whitespace from the hardware name file data.
		trimmedName := strings.Trim(string(nameValueOfHardwareDevice), " \n")

		sensors, err := GetSensorData(trimmedName, dir.Name())
		if err != nil || len(sensors) < 1 {
			debug("Warning: " + dir.Name() + " does not contain " +
				"valid sensor data in the hardware input file, " +
				"ergo no temperature data to print for this device.")

			fmt.Println(
				dir.Name(),
				"-",
				trimmedName,
				"\n└─ no data found",
				"\n")
			continue
		}

		fmt.Print(dir.Name(), " - ", trimmedName, "\n│")

		numberOfSensors := len(sensors)

		for i, sensor := range sensors {

			sensorType := ""
			sensorUnits := ""

			switch sensor.category {
			case "temp":

				// Usually hardware sensors uses 3-sigma of precision and stores
				// the value as an integer for purposes of simplicity.
				//
				// Ergo, this needs to be divided by 1000 to give temperature
				// values that are meaningful to humans.
				//
				sensor.intData /= 1000

				// This acts as a work-around for the k10temp sensor module.
				if sensor.name == "k10temp" &&
					!digitalAmdPowerModuleInUse {

					// Add 30 degrees to the current temperature.
					sensor.intData += 30
				}

				sensorType = "temperature sensor " + strconv.Itoa(sensor.number)
				sensorUnits = "C"

			case "fan":
				sensorType = "fan sensor " + strconv.Itoa(sensor.number)
				sensorUnits = "RPM"

			case "default":
				// assume a default of no units
			}

			if i == 0 && numberOfSensors == 1 {
				fmt.Println(
					"\n├─"+sensorType,
					"\n└─",
					sensor.intData,
					sensorUnits,
					"\n")
			} else if i == numberOfSensors-1 {
				fmt.Println(
					"\n├─"+sensorType,
					"\n└─",
					sensor.intData,
					sensorUnits,
					"\n")
			} else if i == 0 {
				fmt.Print(
					"\n├─"+sensorType,
					"\n├─ ",
					sensor.intData,
					" "+sensorUnits,
					"\n│")
			} else {
				fmt.Print(
					"\n├─"+sensorType,
					"\n├─ ",
					sensor.intData,
					" "+sensorUnits,
					"\n│")
			}
		}
	}
}
