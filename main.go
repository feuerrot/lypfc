package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

type Colorize func(int, int) (int, int, int, bool)

type Sprite struct {
	StartX int
	StartY int
	SizeX  int
	SizeY  int
	Data   []struct {
		R uint8
		G uint8
		B uint8
	}
}

func (sp Sprite) Render() string {
	parts := []string{}
	for y := 0; y < sp.SizeY; y += 1 {
		for x := 0; x < sp.SizeX; x += 1 {
			pxl := sp.Data[x*sp.SizeX+y]
			parts = append(parts, fmt.Sprintf("PX %d %d %02X%02X%02X", x+sp.StartX, y+sp.StartY, pxl.R, pxl.G, pxl.B))
		}
	}

	return strings.Join(parts, "\n")
}

type PF struct {
	SizeX   int
	SizeY   int
	Conns   []PFC
	Sprites []Sprite
}

type PFC struct {
	Num    int
	rd     *bufio.Reader
	rs     *bufio.Scanner
	wr     *bufio.Writer
	conn   net.Conn
	pxlbuf []string
}

func (pfc *PFC) cmd(command string) (string, error) {
	num, err := pfc.wr.WriteString(fmt.Sprintf("%s\n", command))
	if err != nil {
		return "", fmt.Errorf("Can't write %s: %v", command, err)
	}
	log.Printf("Wrote %d bytes", num)
	pfc.wr.Flush()

	pfc.rs.Scan()
	err = pfc.rs.Err()
	if err != nil {
		return "", fmt.Errorf("Can't read answer: %v", err)
	}
	out := pfc.rs.Text()

	return out, nil
}

func (pfc *PFC) px(command string) error {
	_, err := pfc.wr.WriteString(fmt.Sprintf("%s\n", command))
	if err != nil {
		return fmt.Errorf("Can't write %s: %v", command, err)
	}
	return nil
}

func (pfc *PFC) Size() (int, int, error) {
	out, err := pfc.cmd("SIZE")
	if err != nil {
		return 0, 0, fmt.Errorf("Can't get size: %v", err)
	}

	parts := strings.Split(out, " ")
	xsize, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("Can't convert %s to int: %v", parts[1], err)
	}
	ysize, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, fmt.Errorf("Can't convert %s to int: %v", parts[2], err)
	}
	log.Printf("Size: %dx%d", xsize, ysize)

	return xsize, ysize, nil
}

func (pf *PF) Connect(host, port string, connections int) error {
	for i := 0; i < connections; i++ {
		c := PFC{Num: i}
		err := c.Connect(host, port)
		if err != nil {
			return err
		}

		if i == 0 {
			pf.SizeX, pf.SizeY, err = c.Size()
			if err != nil {
				return err
			}
		}
		log.Printf("size: %d %d", pf.SizeX, pf.SizeY)

		pf.Conns = append(pf.Conns, c)
	}

	return nil
}

func (pf *PF) Run() {

	ch := make(chan Sprite)
	for _, c := range pf.Conns {
		c := c
		go c.Draw(ch)
	}

	for {
		rand.Shuffle(len(pf.Sprites), func(i, j int) { pf.Sprites[i], pf.Sprites[j] = pf.Sprites[j], pf.Sprites[i] })
		for _, sp := range pf.Sprites {
			ch <- sp
		}
		//log.Printf("loop")
	}

}

func (pf *PF) Fill(cf Colorize) {
	ssize := 64

	for x := 0; ssize*x < pf.SizeX; x++ {
		for y := 0; ssize*y < pf.SizeY; y++ {
			s := Sprite{
				SizeX:  ssize,
				SizeY:  ssize,
				StartX: x * ssize,
				StartY: y * ssize,
				Data: []struct {
					R uint8
					G uint8
					B uint8
				}{},
			}

			doAppend := false
			for xd := x * ssize; xd < (x+1)*ssize; xd++ {
				for yd := y * ssize; yd < (y+1)*ssize; yd++ {
					r, g, b, draw := cf(xd, yd)
					if draw {
						doAppend = true
					}
					s.Data = append(s.Data, struct {
						R uint8
						G uint8
						B uint8
					}{
						R: uint8(r),
						G: uint8(g),
						B: uint8(b),
					})
				}
			}
			if doAppend {
				pf.Sprites = append(pf.Sprites, s)
			}
		}
	}
}

func (pfc *PFC) Connect(host, port string) error {
	log.Printf("Connecting to %s:%s", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return fmt.Errorf("Can't connect to %s:%s: %v", host, port, err)
	}
	log.Printf("Connected")

	pfc.conn = conn
	pfc.rd = bufio.NewReader(conn)
	pfc.rs = bufio.NewScanner(conn)
	pfc.wr = bufio.NewWriter(conn)

	return nil
}

func (pfc *PFC) Pixel(x, y, r, g, b int) error {
	if len(pfc.pxlbuf) < 10000 {
		pfc.pxlbuf = append(pfc.pxlbuf, fmt.Sprintf("PX %d %d %X%X%X", x, y, r%256, g%256, b%256))
		return nil
	}

	err := pfc.px(strings.Join(pfc.pxlbuf, "\n"))
	if err != nil {
		return fmt.Errorf("Can't write pixel: %v", err)
	}

	pfc.pxlbuf = []string{}

	return nil
}

func (pfc *PFC) Draw(queue chan Sprite) {
	for s := range queue {
		cmds := s.Render()
		err := pfc.px(cmds)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
	}
}

func Color1(x, y int) (int, int, int, bool) {
	return x * y % 3, y % (x + 1), (x * y) / (y*x + 1), true
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <host> <port> <image> <xstart> <ystart>", os.Args[0])
	}

	log.Printf("argcount: %d", len(os.Args))
	pf := PF{}
	pf.Connect(os.Args[1], os.Args[2], 8)

	if len(os.Args) == 3 {
		pf.Fill(Color1)
	}
	if len(os.Args) == 4 {
		fillFunc, err := MetaImage(os.Args[3], 0, 0)
		if err != nil {
			log.Fatal(err)
		}
		pf.Fill(fillFunc)
	}
	if len(os.Args) == 6 {
		x, _ := strconv.Atoi(os.Args[4])
		y, _ := strconv.Atoi(os.Args[5])
		fillFunc, err := MetaImage(os.Args[3], x, y)
		if err != nil {
			log.Fatal(err)
		}
		pf.Fill(fillFunc)
	}
	pf.Run()

}
