package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

//
// Globals
//
var (
	// cpu info location, as of kernel 4.4+
	cpuinfoDirectory = "/proc/cpuinfo"

	// Current location of the hardware sensor data, as of kernel 4.4+
	hardwareMonitorDirectory = "/sys/class/hwmon/"

	// Whether or not to print debug messages.
	debugMode = false

	// Attribute file for storing the hardware device name.
	hardwareNameFile = "name"

	// Attribute file for storing the hardware device current temperature.
	prefixes    = []string{"temp", "fan"}
	inputSuffix = "_input"
	labelSuffix = "_label"

	// flag to check whether the AMD digital thermo module is in use
	digitalAmdPowerModuleInUse = false

	// size of the longest hwmonX/name entry string
	maxEntryLength = 0

	// spacer size
	spacerSize = 4

	// Whether or not to print the current version of the program
	printVersion = false

	// default version value
	Version = "0.0"
)

// Initialize the argument input flags.
func init() {

	flag.BoolVar(&printVersion, "version", false,
		"Print the current version of this program and exit.")

	flag.BoolVar(&debugMode, "debug", false,
		"Dump debug output to stdout.")
}

//
// PROGRAM MAIN
//
func main() {

	flag.Parse()

	if printVersion {
		fmt.Println("tempchk v" + Version)
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
		debug(dir.Name() + " --> " +
			hardwareNameFilepathOfGivenDevice)

		// ...check to see if a 'name' file is present inside the directory.
		nameValueOfHardwareDevice, err := ioutil.ReadFile(hardwareNameFilepathOfGivenDevice)
		if err != nil {

			// If debug mode, then print out a message telling the user
			// which device is missing a hardware 'name' file.
			debug("Warning: " + dir.Name() + " does not contain a " +
				"hardware name file. Skipping...")

			// Move on to the next device.
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

			fmt.Println(dir.Name(), " ", trimmedName, "\t\t n/a")
			continue
		}

		for _, sensor := range sensors {

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

				sensorUnits = "C"
				sensorUnits += "\t\ttemperature sensor " + strconv.Itoa(sensor.number)

			case "fan":
				sensorUnits = "RPM"
				sensorUnits += "\tfan sensor " + strconv.Itoa(sensor.number)

			case "default":
				// assume a default of no units
			}

			fmt.Println(dir.Name(), "\t", sensor.name, "\t", sensor.intData, sensorUnits)
		}
	}
}
