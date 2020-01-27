package main

import (
	"bytes"
	"fmt"
	"github.com/kbinani/screenshot"
	"image"
	"time"
)

const CerealPixels = 64 // must match value in addon lua
const DisplayNumber = 0 // which display you have WoW open on (in fullscreen!)
const Interval = 100 * time.Millisecond

func main() {
	DC := DeCerealizer{}
	stringChan := make(chan string, 64)
	defer close(stringChan)
	go DC.Start(stringChan)
	// just print new messages as they come in
	for {
		str := <-stringChan
		fmt.Println(str)
	}
}

type DeCerealizer struct {
	buf            bytes.Buffer
	expectedMsgNum int
}

func (d *DeCerealizer) Start(out chan<- string) {
	for {
		// align loop with system clock on the interval
		// so with a 100ms interval the snapshot will happen at Xs:0ms, Xs:100ms, Xs:200ms, etc
		// (unless it takes longer than the interval, in which case it'll skip a beat)
		timeUntilAlign := time.Until(time.Now().Add(Interval).Truncate(Interval))
		time.Sleep(timeUntilAlign)
		msgNum, msgBytes, err := d.getNextMsg()
		if _, ok := err.(*ChecksumError); ok {
			//fmt.Println(e.Error())
			continue
		} else if err != nil {
			fmt.Println("error: ", err.Error())
			continue
		}
		if msgNum == d.expectedMsgNum-1 {
			// message has not updated yet
			continue
		} else {
			// TODO: perhaps warn when badly out of sync
			d.expectedMsgNum = msgNum
		}
		for _, b := range msgBytes {
			if b == '\n' {
				// newline means end of string, so send it thru the channel and reset buffer
				out <- d.buf.String()
				d.buf.Reset()
			} else {
				// otherwise keep filling buffer
				d.buf.WriteByte(b)
			}
		}
		d.expectedMsgNum++
	}
}
func (d *DeCerealizer) getNextMsg() (msgNum int, msgBytes []byte, err error) {
	displayBounds := screenshot.GetDisplayBounds(DisplayNumber)
	// bottom right strip of displayBounds
	bounds := image.Rectangle{
		Min: displayBounds.Max.Sub(image.Point{
			X: CerealPixels,
			Y: 1,
		}),
		Max: displayBounds.Max,
	}

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return 0, nil, fmt.Errorf("error capturing screen: %w", err)
	}
	var imgBytes [CerealPixels * 3]byte
	for x := 0; x < CerealPixels; x++ {
		col := img.RGBAAt(x, 0)
		imgBytes[x*3] = col.R
		imgBytes[x*3+1] = col.G
		imgBytes[x*3+2] = col.B
	}
	msgNum = int(imgBytes[0])
	remoteChecksum := int(imgBytes[1])
	msg := imgBytes[2:]

	localChecksum := 0
	for _, b := range msg {
		localChecksum += int(b)
	}
	localChecksum = localChecksum & 0xFF

	if remoteChecksum != localChecksum {
		return 0, nil, &ChecksumError{
			expected: remoteChecksum,
			got:      localChecksum,
			buffer:   imgBytes[:],
		}
	}
	return msgNum, msg, nil
}

type ChecksumError struct {
	expected int
	got      int
	buffer   []byte
}

func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum error (%d != expected %d)", e.got, e.expected)
}
