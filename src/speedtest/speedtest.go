package main
import (
	"fmt"
	"log"
	"time"
	"os"
	"strings"
	"math/rand"
	"flag"
)

import (
	"speedtest/misc"
	"speedtest/debug"
	"speedtest/sthttp"
)

var VERSION = "0.04"

var NUMCLOSEST int
var NUMLATENCYTESTS int
var TESTSERVERID = ""


func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.BoolVar(&debug.DEBUG, "d", false, "Turn on debugging")
	listFlag := flag.Bool("l", false, "List servers (hint use 'grep' or 'findstr' to locate a server ID to use for '-s'")
	flag.BoolVar(&debug.QUIET, "q", false, "Quiet Mode. Only output server and results")
	flag.StringVar(&TESTSERVERID, "s", "", "Specify a server ID to use")
	flag.IntVar(&NUMCLOSEST, "nc", 3, "Number of geographically close servers to test to find the optimal server")
	flag.IntVar(&NUMLATENCYTESTS, "nl", 3, "Number of latency tests to perform to determine which server is the fastest")
	verFlag := flag.Bool("v", false, "Display version")
	
	flag.Parse()
    
	if *verFlag == true {
		fmt.Printf("%s - Version: %s\n", os.Args[0], VERSION)
		os.Exit(0)
	}
	
	if debug.DEBUG { 
		log.Printf("Debugging on...\n") 
		debug.QUIET = false
	}
	
	if *listFlag == true {
		if debug.DEBUG { fmt.Printf("Loading config from speedtest.net\n") }
		sthttp.CONFIG = sthttp.GetConfig()
		
		if debug.DEBUG { fmt.Printf("Getting servers list...") }
		allServers := sthttp.GetServers()
		if debug.DEBUG { fmt.Printf("(%d) found\n", len(allServers)) }
		for s := range allServers {
			server := allServers[s]
			printServer(server)
		}
		os.Exit(0)
	}
}

// FIXME: need to do some more work here to make sure this is optimal
// https://github.com/zpeters/speedtest/issues/6
func downloadTest(server sthttp.Server) float64 {
	var urls []string
	var speedAcc float64

	//dlsizes := []int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}
	dlsizes := []int{350, 500, 750, 1000, 1500, 2000}
	numRuns := 1

	for size := range dlsizes {
		for i := 0; i<numRuns; i++ {
			url := server.Url
			splits := strings.Split(url, "/")
			baseUrl := strings.Join(splits[1:len(splits) -1], "/")
			randomImage := fmt.Sprintf("random%dx%d.jpg", dlsizes[size], dlsizes[size])
			downloadUrl := "http:/" + baseUrl + "/" + randomImage
			urls = append(urls, downloadUrl)
		}
	}	


	if !debug.QUIET { fmt.Printf("Testing download speed") }
	if debug.DEBUG { fmt.Printf("\n") }

	for u := range urls {

		dlSpeed := sthttp.DownloadSpeed(urls[u])
		if debug.DEBUG { 
			fmt.Printf("Download test %d\n", u) 
		} else if !debug.QUIET {
			 fmt.Printf(".")
		 }
		speedAcc = speedAcc + dlSpeed
	}
	
	if !debug.QUIET { fmt.Printf("\n") }

	mbps := (speedAcc / float64(len(urls)))
	return mbps
}


// FIXME: need to generally do some more work in here
// https://github.com/zpeters/speedtest/issues/7
func uploadTest(server sthttp.Server) float64 {
	// https://github.com/sivel/speedtest-cli/blob/master/speedtest-cli
	var ulsize []int
	var ulSpeedAcc float64

	ulsizesizes := []int{int(0.25 * 1024 * 1024), int(0.5 * 1024 * 1024)}

	var numRuns = 3
	
	for size := range ulsizesizes {
		for i := 0; i<numRuns; i++ {
			ulsize = append(ulsize, ulsizesizes[size])
		}
	}

	if !debug.QUIET { fmt.Printf("Testing upload speed") }
	if debug.DEBUG { fmt.Printf("\n") }
	
	for i:=0; i<len(ulsize); i++ {

		r := misc.Urandom(ulsize[i])
		ulSpeed := sthttp.UploadSpeed(server.Url, "text/xml", r)
		if debug.DEBUG { 
			fmt.Printf("Ulsize: %d\n", ulsize[i]) 
		} else if !debug.QUIET {
			 fmt.Printf(".")
		 }
		ulSpeedAcc = ulSpeedAcc + ulSpeed
	}
	
	if !debug.QUIET { fmt.Printf("\n") }

	mbps := ulSpeedAcc / float64(len(ulsize))
	return mbps
}

func findServer(id string, serversList []sthttp.Server) sthttp.Server {
	var foundServer sthttp.Server
	for s := range serversList {
		if serversList[s].Id == id {
			foundServer = serversList[s]
		}
	}
	if foundServer.Id == "" {
		log.Panicf("Cannot locate server Id '%s' in our list of speedtest servers!\n", id)
	}
	return foundServer
}

func printServer(server sthttp.Server) {
    fmt.Printf("%-4s | %s (%s, %s)\n", server.Id, server.Sponsor, server.Name, server.Country)
}

func main() {
	var testServer sthttp.Server

	if debug.DEBUG { fmt.Printf("Loading config from speedtest.net\n") }
	sthttp.CONFIG = sthttp.GetConfig()
	
	if debug.DEBUG { fmt.Printf("Getting servers list...") }
	allServers := sthttp.GetServers()
	if debug.DEBUG { fmt.Printf("(%d) found\n", len(allServers)) }
    
	if TESTSERVERID != "" {		
		// they specified a server so find it in the list
		testServer = findServer(TESTSERVERID, allServers)
		printServer(testServer)
		if !debug.QUIET { fmt.Printf("Testing latency...\n") }
		testServer.AvgLatency = sthttp.GetLatency(testServer, NUMLATENCYTESTS)
	} else {
		// find a fast server for them
		closestServers := sthttp.GetClosestServers(NUMCLOSEST, allServers)
		if !debug.QUIET { fmt.Printf("Finding fastest server..\n") }
		testServer = sthttp.GetFastestServer(NUMLATENCYTESTS, closestServers)
		printServer(testServer)
	}

	dmbps := downloadTest(testServer)	
	umbps := uploadTest(testServer)
	
	fmt.Printf("Ping: %s | Download: %3.2f Mbps | Upload: %3.2f Mbps\n", testServer.AvgLatency, dmbps, umbps)
}

