package tests

import (
	"fmt"
	"log"
	"strings"
)

import (
	"github.com/pespin/speedtest/debug"
	"github.com/pespin/speedtest/misc"
	"github.com/pespin/speedtest/print"
	"github.com/pespin/speedtest/settings"
	"github.com/pespin/speedtest/sthttp"
)

func triggerDwlOverflow(bigurl string) {
	var dlSpeed float64
	var beforeSpeed float64

	beforeSpeed = sthttp.DownloadSpeed(bigurl)
	if debug.DEBUG {
		log.Printf("Initial download speed: %f mbps", beforeSpeed)
	}

	for ok := true; ok; {
		dlSpeed = sthttp.DownloadSpeed(bigurl)
		if dlSpeed < beforeSpeed {
			break;
		} else {
			if debug.DEBUG {
				log.Printf("New improved download speed: %f mbps", dlSpeed)
			}
			beforeSpeed = dlSpeed
		}
	}

	if debug.DEBUG {
		log.Printf("Final max download speed: %f mbps (it dropped to %f)", beforeSpeed, dlSpeed)
	}

}

func triggerUplOverflow(server sthttp.Server, bigsize int) {
	var ulSpeed float64
	var beforeSpeed float64

	r := misc.Urandom(bigsize)
	beforeSpeed = sthttp.UploadSpeed(server.URL, "text/xml", r)
	if debug.DEBUG {
		log.Printf("Initial upload speed: %f mbps", beforeSpeed)
	}

	for ok := true; ok; {
		ulSpeed = sthttp.UploadSpeed(server.URL, "text/xml", r)
		if ulSpeed < beforeSpeed {
			break;
		} else {
			if debug.DEBUG {
				log.Printf("New improved upload speed: %f mbps", ulSpeed)
			}
			beforeSpeed = ulSpeed
		}
	}

	if debug.DEBUG {
		log.Printf("Final max upload speed: %f mbps (it dropped to %f)", beforeSpeed, ulSpeed)
	}

}

// DownloadTest will perform the "normal" speedtest download test
func DownloadTest(server sthttp.Server) float64 {
	var urls []string
	var maxSpeed float64
	var avgSpeed float64

	// http://speedtest1.newbreakcommunications.net/speedtest/speedtest/
	dlsizes := []int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}

	for size := range dlsizes {
		url := server.URL
		splits := strings.Split(url, "/")
		baseURL := strings.Join(splits[1:len(splits)-1], "/")
		randomImage := fmt.Sprintf("random%dx%d.jpg", dlsizes[size], dlsizes[size])
		downloadURL := "http:/" + baseURL + "/" + randomImage
		urls = append(urls, downloadURL)
	}

	if !debug.QUIET && !debug.REPORT {
		log.Printf("Testing download speed")
	}

	var bigurl = urls[len(urls)-1]
	if debug.DEBUG {
		log.Printf("Downloading some data to trigger overflow (%s)", bigurl)
	}
	triggerDwlOverflow(bigurl)

	for u := range urls {

		if debug.DEBUG {
			log.Printf("Download Test Run: %s\n", urls[u])
		}
		dlSpeed := sthttp.DownloadSpeed(urls[u])
		if !debug.QUIET && !debug.DEBUG && !debug.REPORT {
			fmt.Printf(".")
		}
		if debug.DEBUG {
			log.Printf("Dl Speed: %v\n", dlSpeed)
		}

		if settings.ALGOTYPE == "max" {
			if dlSpeed > maxSpeed {
				maxSpeed = dlSpeed
			}
		} else {
			avgSpeed = avgSpeed + dlSpeed
		}

	}

	if !debug.QUIET && !debug.REPORT {
		fmt.Printf("\n")
	}

	if settings.ALGOTYPE != "max" {
		return avgSpeed / float64(len(urls))
	}
	return maxSpeed

}

// UploadTest runs a "normal" speedtest upload test
func UploadTest(server sthttp.Server) float64 {
	// https://github.com/sivel/speedtest-cli/blob/master/speedtest-cli
	var ulsize []int
	var maxSpeed float64
	var avgSpeed float64

	ulsizesizes := []int{
		int(0.25 * 1024 * 1024),
		int(0.5 * 1024 * 1024),
		int(1.0 * 1024 * 1024),
		int(1.5 * 1024 * 1024),
		int(2.0 * 1024 * 1024),
	}

	for size := range ulsizesizes {
		ulsize = append(ulsize, ulsizesizes[size])
	}

	if !debug.QUIET && !debug.REPORT {
		log.Printf("Testing upload speed")
	}

	var bigsize = ulsizesizes[len(ulsizesizes)-1]
	if debug.DEBUG {
		log.Printf("Uploading some data to trigger overflow (size %d)", bigsize)
	}
	triggerUplOverflow(server, bigsize)

	for i := 0; i < len(ulsize); i++ {
		if debug.DEBUG {
			log.Printf("Upload Test Run: %v\n", i)
		}
		r := misc.Urandom(ulsize[i])
		ulSpeed := sthttp.UploadSpeed(server.URL, "text/xml", r)
		if !debug.QUIET && !debug.DEBUG && !debug.REPORT {
			fmt.Printf(".")
		}

		if settings.ALGOTYPE == "max" {
			if ulSpeed > maxSpeed {
				maxSpeed = ulSpeed
			}
		} else {
			avgSpeed = avgSpeed + ulSpeed
		}

	}

	if !debug.QUIET && !debug.REPORT {
		fmt.Printf("\n")
	}

	if settings.ALGOTYPE != "max" {
		return avgSpeed / float64(len(ulsizesizes))
	}
	return maxSpeed
}

// FindServer will find a specific server in the servers list
func FindServer(id string, serversList []sthttp.Server) sthttp.Server {
	var foundServer sthttp.Server
	for s := range serversList {
		if serversList[s].ID == id {
			foundServer = serversList[s]
		}
	}
	if foundServer.ID == "" {
		log.Fatalf("Cannot locate server Id '%s' in our list of speedtest servers!\n", id)
	}
	return foundServer
}

// ListServers prints a list of all "global" servers
func ListServers() {
	if debug.DEBUG {
		fmt.Printf("Loading config from speedtest.net\n")
	}
	sthttp.CONFIG = sthttp.GetConfig()
	if debug.DEBUG {
		fmt.Printf("\n")
	}

	if debug.DEBUG {
		fmt.Printf("Getting servers list...")
	}
	allServers := sthttp.GetServers()
	if debug.DEBUG {
		fmt.Printf("(%d) found\n", len(allServers))
	}
	for s := range allServers {
		server := allServers[s]
		print.Server(server)
	}
}
