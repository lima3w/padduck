package scanner

import (
	"context"
	"net"
	"sync"
	"time"
)

// ScanPorts performs a TCP connect scan against the given IP for each port in the
// ports slice. It runs at most concurrency probes simultaneously. A port is marked
// open if a TCP connection succeeds within timeout.
func ScanPorts(ctx context.Context, ip string, ports []string, concurrency int, timeout time.Duration) map[string]bool {
	if concurrency <= 0 {
		concurrency = 10
	}
	if timeout <= 0 {
		timeout = time.Second
	}

	type work struct{ port string }
	type result struct {
		port string
		open bool
	}

	workCh := make(chan work, len(ports))
	resultCh := make(chan result, len(ports))

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range workCh {
				select {
				case <-ctx.Done():
					resultCh <- result{port: w.port, open: false}
					continue
				default:
				}
				addr := net.JoinHostPort(ip, w.port)
				conn, err := net.DialTimeout("tcp", addr, timeout)
				open := err == nil
				if open {
					conn.Close()
				}
				resultCh <- result{port: w.port, open: open}
			}
		}()
	}

	for _, p := range ports {
		workCh <- work{port: p}
	}
	close(workCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	out := make(map[string]bool, len(ports))
	for r := range resultCh {
		out[r.port] = r.open
	}
	return out
}
