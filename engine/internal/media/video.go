package media

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"sync"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// VideoFrame represents a synthesized video frame
type VideoFrame struct {
	FrameIndex int64     `json:"frame_index"`
	Timestamp  time.Time `json:"timestamp"`
	Codec      string    `json:"codec"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	DataB64    string    `json:"data_base64"`
	SizeBytes  int       `json:"size_bytes"`
}

// VideoStreamGenerator simulates real-time CCTV, optical sensor, or drone camera feeds
type VideoStreamGenerator struct {
	mu         sync.Mutex
	cfg        *pb.VideoStreamFunction
	frameIndex int64
	startTime  time.Time
}

// NewVideoStreamGenerator creates a video stream frame broadcaster
func NewVideoStreamGenerator(cfg *pb.VideoStreamFunction) *VideoStreamGenerator {
	if cfg == nil {
		cfg = &pb.VideoStreamFunction{
			Codec:           "mjpeg",
			Width:           320,
			Height:          240,
			Fps:             30,
			Pattern:         "bouncing_ball",
			RenderTimestamp: true,
		}
	}
	if cfg.Width <= 0 {
		cfg.Width = 320
	}
	if cfg.Height <= 0 {
		cfg.Height = 240
	}
	if cfg.Fps <= 0 {
		cfg.Fps = 30
	}

	return &VideoStreamGenerator{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// NextFrame synthesizes the next continuous video frame
func (g *VideoStreamGenerator) NextFrame() (*VideoFrame, error) {
	g.mu.Lock()
	idx := g.frameIndex
	g.frameIndex++
	g.mu.Unlock()

	w := int(g.cfg.Width)
	h := int(g.cfg.Height)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Draw pattern
	switch g.cfg.Pattern {
	case "color_bars":
		drawColorBars(img, w, h)
	case "noise":
		drawSensorNoise(img, w, h, idx)
	case "bouncing_ball":
		fallthrough
	default:
		drawBouncingBall(img, w, h, idx)
	}

	// Compress frame to JPEG (MJPEG container frame)
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, fmt.Errorf("failed to encode video frame: %w", err)
	}

	frameBytes := buf.Bytes()
	frameTime := g.startTime.Add(time.Duration(idx*int64(time.Second)) / time.Duration(g.cfg.Fps))

	return &VideoFrame{
		FrameIndex: idx,
		Timestamp:  frameTime,
		Codec:      g.cfg.Codec,
		Width:      w,
		Height:     h,
		DataB64:    base64.StdEncoding.EncodeToString(frameBytes),
		SizeBytes:  len(frameBytes),
	}, nil
}

// drawBouncingBall simulates motion trajectory for computer vision object tracking
func drawBouncingBall(img *image.RGBA, w, h int, frameIdx int64) {
	// Dark background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{20, 24, 33, 255}}, image.Point{}, draw.Src)

	// Ball motion coordinates (harmonic oscillation)
	speed := float64(frameIdx) * 0.1
	radius := int(math.Min(float64(w), float64(h)) * 0.08)
	if radius < 4 {
		radius = 4
	}

	cx := int(float64(w)/2.0 + (float64(w)/2.0-float64(radius)-10.0)*math.Sin(speed))
	cy := int(float64(h)/2.0 + (float64(h)/2.0-float64(radius)-10.0)*math.Abs(math.Cos(speed*1.3)))

	ballColor := color.RGBA{0, 220, 130, 255} // Emerald camera target

	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				dx := x - cx
				dy := y - cy
				if dx*dx+dy*dy <= radius*radius {
					img.Set(x, y, ballColor)
				}
			}
		}
	}
}

// drawColorBars renders standard SMPTE-style test color bars
func drawColorBars(img *image.RGBA, w, h int) {
	bars := []color.RGBA{
		{192, 192, 192, 255}, // Gray
		{192, 192, 0, 255},   // Yellow
		{0, 192, 192, 255},   // Cyan
		{0, 192, 0, 255},     // Green
		{192, 0, 192, 255},   // Magenta
		{192, 0, 0, 255},     // Red
		{0, 0, 192, 255},     // Blue
	}
	barWidth := w / len(bars)
	for i, c := range bars {
		r := image.Rect(i*barWidth, 0, (i+1)*barWidth, h)
		draw.Draw(img, r, &image.Uniform{c}, image.Point{}, draw.Src)
	}
}

// drawSensorNoise simulates optical sensor noise / thermal camera static
func drawSensorNoise(img *image.RGBA, w, h int, frameIdx int64) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Pseudo-random deterministic noise hash
			v := uint8((x*374761393 + y*668265263 + int(frameIdx)*3266489917) % 256)
			img.Set(x, y, color.RGBA{v, v, v, 255})
		}
	}
}
