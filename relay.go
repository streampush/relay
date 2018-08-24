package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/nareix/joy4/av/pubsub"

	"github.com/nareix/joy4/av/avutil"

	// "github.com/nareix/joy4/avutil"
	// "github.com/nareix/joy4/format"
	"github.com/nareix/joy4/format/rtmp"

	"github.com/abiosoft/ishell"
)

const _CfgPort = "1935"

var restreams = struct {
	sync.RWMutex
	cfgs map[string]*Restream
}{cfgs: make(map[string]*Restream)}

func reloadConfigs() {
	log.Println("Reloading restream configs")

	files, err := ioutil.ReadDir("configs")
	if err != nil {
		shell.Println("Error loading restream configs", err)
		return
	}

	for _, file := range files {
		cfgFile, err := os.Open(path.Join("configs", file.Name()))
		if err != nil {
			shell.Println("Error loading config", err)
			continue
		}

		bytes, _ := ioutil.ReadAll(cfgFile)
		toInsert := &Restream{}
		err = json.Unmarshal(bytes, toInsert)
		cfgFile.Close()
		if err != nil {
			shell.Println("Error loading config", err)
			continue
		}

		restreams.RLock()
		restream, exists := restreams.cfgs[toInsert.ID]
		restreams.RUnlock()

		restreams.Lock()
		if exists {
			log.Printf("^ Updating restream: %s\n", restream.Name)
			restream.Name = toInsert.Name

			/* Find new endpoints and endpoints that need to be
			 * updated in the old config from the new config */
			for newEndpointID, newEndpoint := range toInsert.Endpoints {
				oldEndpoint, exists := restream.Endpoints[newEndpointID]
				if exists { // Update the endpoint
					log.Printf("\t^ Updating endpoint: %s\n", oldEndpoint.Name)
					oldEndpoint.Update(*newEndpoint)

					// TODO: Should check if URL changed here; if it did
					// restart endpoint stream. If not, do nothing.
				} else { // Insert the new endpoint
					log.Printf("\t+ Adding endpoint: %s\n", newEndpoint.Name)
					restream.Endpoints[newEndpointID] = newEndpoint

					// If this endpoint is currently live, make sure we
					// push to our new endpoints
					if restream.Streaming {
						go pushStream(restream, newEndpoint)
					}
				}
			}

			/* Find endpoints from the old config that don't exist
			 * in the new config */
			for oldEndpointID, oldEndpoint := range restream.Endpoints {
				_, exists := toInsert.Endpoints[oldEndpointID]
				if !exists {
					oldEndpoint.Dest.Close()
					delete(restream.Endpoints, oldEndpointID)
					log.Printf("\t- Removed old endpoint %s", oldEndpoint.Name)
				}
			}

			select {
			case restream.Channel <- "reload": // Hot-reload endpoints
				break
			default:
				break
			}
		} else {
			log.Printf("+ New restream: %s\n", toInsert.Name)
			toInsert.Channel = make(chan string)
			restreams.cfgs[toInsert.ID] = toInsert
		}
		restreams.Unlock()
	}

	log.Printf("Loaded %d configs\n", len(restreams.cfgs))
}

func pushStream(restream *Restream, endpoint *Endpoint) {
	// Open connection to destination
	dest, err := rtmp.Dial(endpoint.URL)
	if err != nil {
		shell.Println("Error restreaming", err)
		return
	}

	endpoint.Dest = dest
	endpoint.Connected = true
	endpoint.ConnErr = nil

	// Write header to destination
	streams, _ := restream.Origin.Streams()
	dest.WriteHeader(streams)

	// Copy packets from origin queue to destination
	go func() {
		err := avutil.CopyPackets(dest, restream.Queue.Latest())
		if err != nil {
			endpoint.ConnErr = err
		}
		dest.WriteTrailer()
		endpoint.Connected = false
	}()

	log.Printf("Pushing to %s", endpoint.Name)

	// waitChan:
	// 	for {
	// 		switch <-restream.Channel {
	// 		case "quit":
	// 			shell.Println("Stopping push to", endpoint.Name)
	// 			dest.Close()
	// 			break waitChan
	// 		case "reload":
	// 			// TODO: If a endpoint has not changed, do nothing
	// 			// if an endpoint has been removed, stop streaming
	// 			fmt.Println("Endpoints", restream.Endpoints)
	// 			break
	// 		default:
	// 			shell.Println("Invalid message")
	// 		}
	// 	}
}

var shell *ishell.Shell

func main() {
	shell = ishell.New()

	reloadConfigs()

	server := &rtmp.Server{
		Addr: ":" + _CfgPort,
	}

	server.HandlePublish = func(conn *rtmp.Conn) {
		_, err := conn.Streams()
		if err != nil {
			log.Println("Error handling publish:", err)
		}

		params := strings.Split(conn.URL.RequestURI(), "/")
		endpoint := params[1]

		restreams.RLock()
		restream, exists := restreams.cfgs[endpoint]
		restreams.RUnlock()

		if !exists {
			log.Println("Invalid stream ID; dropping connection.")
			conn.Close()
			return
		}

		restream.Origin = conn
		restream.Queue = pubsub.NewQueue()

		go func() {
			// Copy packets to the queue (this is blocking)
			avutil.CopyPackets(restream.Queue, restream.Origin)

			// Origin stopped sending data
			restream.Channel <- "publish_done"
		}()

		go func() {
		chanLoop:
			for {
				switch <-restream.Channel {
				case "quit":
					fallthrough
				case "publish_done":
					log.Println("Dropping restream", restream.Name)
					restream.Origin.Close()
					restream.Queue.Close()
					restream.Streaming = false
					break chanLoop
				case "reload":
					// Configs have been reloaded, we need to determine if
					// we should spin up any new endpoints. EndpointsS that
					// were removed will kill themselves.
					break
				}
			}
		}()

		for _, endpoint := range restream.Endpoints {
			go pushStream(restream, endpoint)
		}

		restream.Streaming = true
	}

	log.Println("Starting server")
	go server.ListenAndServe()
	log.Println("Server started")

	shell.AddCmd(&ishell.Cmd{
		Name: "reload",
		Help: "reload configs",
		Func: func(c *ishell.Context) {
			reloadConfigs()
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "restreams",
		Help: "list restreams",
		Func: func(c *ishell.Context) {
			restreams.RLock()
			for _, restream := range restreams.cfgs {
				c.Printf("%s] %s - streaming: %t\n", restream.ID, restream.Name, restream.Streaming)
				for idx, endpoint := range restream.Endpoints {
					c.Printf("\t%s] %s - pushing: %t\n", idx, endpoint.Name, endpoint.Connected)
					if restream.Streaming && !endpoint.Connected {
						c.Printf("\t\tError: %s\n", endpoint.ConnErr)
					}
				}
			}
			restreams.RUnlock()
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "debug",
		Help: "toggle debug messages",
		Func: func(c *ishell.Context) {
			rtmp.Debug = !rtmp.Debug
		},
	})

	restreamCmdGroup := &ishell.Cmd{
		Name: "restream",
		Help: "restream operations",
	}

	restreamCmdGroup.AddCmd(&ishell.Cmd{
		Name: "stop",
		Help: "stop a restream: stop <restreamId>",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Missing restream ID")
				return
			}

			restreams.RLock()
			restream, exists := restreams.cfgs[c.Args[0]]
			restreams.RUnlock()

			if !exists {
				c.Println("Restream with that ID does not exist.")
				return
			}

			select {
			case restream.Channel <- "quit":
				c.Printf("Stop command sent to %s\n", restream.Name)
				break
			default:
				c.Println("Restream wasn't live.")
				break
			}
		},
	})

	shell.AddCmd(restreamCmdGroup)

	shell.Run()
}
