package main

import (
	"bytes"
	"flag"
	"fmt"
	client "github.com/whyrusleeping/go-tftp/client"
	"io"
	"io/ioutil"
	"runtime"
	"sync"
	"time"
)

var Quiet bool = false

func benchReads(server, file string, threads, loops, blocksize int, reuse bool, verify []byte) {
	wg := &sync.WaitGroup{}

	bwcollect := make(chan int, 32)
	before := time.Now()

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			var err error
			var cli *client.TftpClient
			defer wg.Done()
			if reuse {
				cli, err = client.NewTftpClient(server)
				cli.Blocksize = blocksize
				if err != nil {
					panic(err)
				}
			}

			for j := 0; j < loops; j++ {
				if !reuse {
					cli, err = client.NewTftpClient(server)
					cli.Blocksize = blocksize
					if err != nil {
						panic(err)
					}
				}
				var buf *bytes.Buffer
				var out io.Writer

				if verify != nil {
					buf = new(bytes.Buffer)
					out = buf
				}
				nbytes, err := cli.GetFile(file, out)
				if err != nil {
					panic(err)
				}

				// Check data if we have a verify chunk
				if verify != nil {
					if !bytes.Equal(verify, buf.Bytes()) {
						fmt.Println("ERROR: incorrect data")
					}
				}
				bwcollect <- nbytes
				if !reuse {
					cli.Close()
				}
			}
			if reuse {
				cli.Close()
			}
		}()
	}

	total := threads * loops
	i := 0

	go func() {
		wg.Wait()
		close(bwcollect)
	}()

	sum := 0
	for bw := range bwcollect {
		sum += bw
		i++
		if !Quiet {
			fmt.Printf("\r%d/%d", i, total)
		}
	}
	fmt.Println()
	took := time.Now().Sub(before)

	fmt.Printf("Total Transferred: %d\n", sum)
	fmt.Printf("Overall Bandwidth: %.0f Bps\n", float64(sum)/took.Seconds())
}

func benchWrites(server string, threads, loops, nbytes, blocksize int, reuse bool) {
	wg := &sync.WaitGroup{}

	bwcollect := make(chan int, 32)
	before := time.Now()

	total := threads * loops
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(thr int) {
			defer wg.Done()
			var cli *client.TftpClient
			var err error
			if reuse {
				cli, err = client.NewTftpClient(server)
				cli.Blocksize = blocksize
				if err != nil {
					panic(err)
				}
			}

			for j := 0; j < loops; j++ {
				if !reuse {
					cli, err = client.NewTftpClient(server)
					cli.Blocksize = blocksize
					if err != nil {
						panic(err)
					}

				}
				read := io.LimitReader(DataReader{}, int64(nbytes))
				nbytes, err := cli.PutFile(fmt.Sprintf("file%d-%d", thr, j), read)
				if err != nil {
					panic(err)
				}
				bwcollect <- nbytes
				if !reuse {
					cli.Close()
				}
			}
			if reuse {
				cli.Close()
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(bwcollect)
	}()

	sum := 0
	i := 0
	for bw := range bwcollect {
		sum += bw
		i++
		if !Quiet {
			fmt.Printf("\r%d/%d", i, total)
		}
	}
	fmt.Println()
	took := time.Now().Sub(before)

	fmt.Printf("Total Transferred: %d\n", sum)
	fmt.Printf("Overall Bandwidth: %.0f Bps\n", float64(sum)/took.Seconds())
}

func main() {
	nprocs := flag.Int("procs", 1, "number of procs to run")
	nthreads := flag.Int("threads", 1, "number of threads to run")
	nloops := flag.Int("loops", 1, "number of operations per thread")
	serv := flag.String("serv", "127.0.0.1:6900", "address of server to benchmark")
	filename := flag.String("file", "", "name of file to work with (for reads only)")
	verify := flag.String("verify", "", "name of local file to check against (for reads only)")
	upload := flag.Int("upload", -1, "size of data for upload testing")
	blocksize := flag.Int("blocksize", 512, "tftp blocksize")
	reuseport := flag.Bool("reuseport", true, "whether or not to reuse the same ports")
	quiet := flag.Bool("quiet", false, "quiet extraneous output")

	flag.Parse()

	Quiet = *quiet
	runtime.GOMAXPROCS(*nprocs)

	fmt.Printf("Testing Server: '%s'\n", *serv)

	if *upload > 0 {
		benchWrites(*serv, *nthreads, *nloops, *upload, *blocksize, *reuseport)
	} else {
		var veriData []byte
		if len(*verify) > 0 {
			out, err := ioutil.ReadFile(*verify)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				return
			}
			veriData = out
		}
		benchReads(*serv, *filename, *nthreads, *nloops, *blocksize, *reuseport, veriData)
	}
}
