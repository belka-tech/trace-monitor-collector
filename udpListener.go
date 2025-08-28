package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"trace-monitor-collector/command"
	"trace-monitor-collector/config"
	"trace-monitor-collector/counter"
	"trace-monitor-collector/traceCollection"
)

var (
	lastPackage         string
	totalPackagesCaught counter.CounterStruct
	totalPackagesParse  counter.CounterStruct
	totalChannelReset   counter.CounterStruct
	channelList         []chan []byte
	start               time.Time
	udpServerReadyChan  = make(chan struct{})
)

func handleUdp(ctx context.Context, cfg *config.Config) {
	defer recoverRoutineHandleUdp(ctx, cfg)

	channelList = make([]chan []byte, cfg.UdpPortRangeCount)

	for port := cfg.UdpPortStart; port <= cfg.UdpPortEnd; port++ {
		localPort := port
		localChannelKey := port - cfg.UdpPortStart

		channelList[localChannelKey] = make(chan []byte, cfg.PacketsSize)
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", localPort))
		if err != nil {
			log.Printf("Error resolving UDP address: %v", err)
			os.Exit(1)
		}
		udpConn, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Printf("Error listening on UDP port %d: %v", localPort, err)
			os.Exit(1)
		}
		defer udpConn.Close()

		if cfg.IsVerboseByLevel("v") {
			log.Println("UDP listener started on", localPort)
		}
		go channelWriter(cfg, localChannelKey, udpConn)
	}

	channelReader(cfg)

	udpServerReadyChan <- struct{}{}

	if cfg.IsVerboseByLevel("vv") {
		var lastValue uint64 = 0
		var clearIter int = 0
		var startTotalPackagesCaught uint64 = 0
		var startTotalPackagesParse uint64 = 0
		for {
			localTotalPackagesCaught := totalPackagesCaught.Count() - startTotalPackagesCaught
			localTotalPackagesParse := totalPackagesParse.Count() - startTotalPackagesParse
			if lastValue == localTotalPackagesCaught {
				clearIter++
				if clearIter >= 5 {
					clearIter = 0
					startTotalPackagesCaught = totalPackagesCaught.Count()
					startTotalPackagesParse = totalPackagesParse.Count()
				}
			} else {
				elapsed := time.Since(start).Milliseconds()
				rps := (float64(localTotalPackagesCaught) / float64(elapsed)) * 1000
				log.Println(localTotalPackagesCaught, " / ", localTotalPackagesParse, " : ", elapsed)
				log.Printf("RPS: %.2f\n", rps)
				clearIter = 0
			}

			lastValue = localTotalPackagesCaught
			time.Sleep(10 * time.Millisecond)
		}
	} else {
		for {
			select {
			case <-ctx.Done():
				totalPackagesCaught.Reset()
				totalPackagesParse.Reset()
				traceCollection.TotalAllSpanClose.Reset()
				traceCollection.TotalTraceSet.Reset()
				traceCollection.TotalSpanSet.Reset()
				traceCollection.TotalTraceDelete.Reset()
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}
}

func channelWriter(cfg *config.Config, localChannelKey int, udpConn *net.UDPConn) {
	buffer := make([]byte, cfg.Buffer)
	for {
		n, _, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if cfg.IsVerboseByLevel("vv") {
			if totalPackagesCaught.Count() == 0 {
				start = time.Now()
			}
			cmd, _ := command.FromJson(buffer[:n])
			log.Println("write:",
				localChannelKey,
				cmd.Pid,
				cmd.Method,
				cmd.TraceId,
				cmd.SentAt)
		}
		if cfg.IsVerboseByLevel("vvv") {
			log.Println("packet:", localChannelKey, string(buffer[:n]))
		}

		packet := make([]byte, n)
		copy(packet, buffer[:n])

		if err := pushToChannal(localChannelKey, packet); err != nil {
			if cfg.IsVerboseByLevel("v") {
				log.Println("_warn:", localChannelKey, err)
			}
		}

		totalPackagesCaught.Increment()
	}
}

func pushToChannal(channelKey int, packet []byte) error {
	select {
	case channelList[channelKey] <- packet:
		return nil
	default:
		for {
			select {
			case <-channelList[channelKey]:
			default:
				channelList[channelKey] <- packet
				totalChannelReset.Increment()
				return fmt.Errorf("chanel reset")
			}
		}
	}
}

func channelReader(cfg *config.Config) {
	for channelKey, channel := range channelList {
		localChannelKey := channelKey
		go func(localChannelKey int, channel chan []byte) {
			for packet := range channel {
				processUdpPacket(cfg, localChannelKey, packet)
			}
		}(localChannelKey, channel)
	}
}

func processUdpPacket(cfg *config.Config, channelKey int, packet []byte) {
	defer recoverPackageProcess()

	cmd, err := command.FromJson(packet)
	if err != nil {
		// TODO: log or return
	}

	if err := applyUdpCommand(cfg, channelKey, cmd); err != nil {
		if cfg.IsVerboseByLevel("v") {
			log.Println("_warn:", channelKey, err)
		}
	}

	totalPackagesParse.Increment()
}

func applyUdpCommand(cfg *config.Config, channelKey int, cmd command.Command) error {
	if cfg.IsVerboseByLevel("vv") {
		log.Println("_read:", channelKey, cmd.Pid, cmd.Method, cmd.TraceId, cmd.SentAt)
	}
	if cfg.IsVerboseByLevel("vvv") {
		log.Println("_data:", channelKey, string(cmd.Data))
	}
	if cmd.Method == "init-trace" {
		return traceCollection.InitTrace(cfg, cmd.Pid, cmd.TraceId, cmd.SentAt, cmd.RawCommand)
	}
	if cmd.Method == "set-trace-current-span" {
		if cmd.Data == nil {
			return traceCollection.DeleteSpan(cfg, cmd.Pid, cmd.TraceId, cmd.SentAt)
		} else {
			return traceCollection.SetTraceCurrentSpan(cfg, cmd.Pid, cmd.TraceId, cmd.SentAt, cmd.RawCommand)
		}
	}
	if cmd.Method == "free-pid" {
		return traceCollection.DeleteTrace(cfg, cmd.Pid, cmd.TraceId, cmd.SentAt)
	}

	return fmt.Errorf("unknown method specified in UDP packet. %v", cmd.Method)
}

func recoverRoutineHandleUdp(ctx context.Context, cfg *config.Config) {
	if r := recover(); r != nil {
		log.Println("Handle UDP error: ", r)
		if cfg.IsVerboseByLevel("v") {
			log.Println("Last package: ", lastPackage)
		}
		go handleUdp(ctx, cfg)
	}
}

func recoverPackageProcess() {
	if r := recover(); r != nil {
		log.Println("Package process error: ", r)
	}
}
