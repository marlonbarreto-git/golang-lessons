// Package main - Chapter 093: Images and Audio Processing
// Go has native image support in stdlib and libraries for advanced
// audio processing including playback, recording, and DSP.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
)

func main() {
	fmt.Println("=== IMAGES AND AUDIO PROCESSING IN GO ===")

	// ============================================
	// PART 1: IMAGE PROCESSING
	// ============================================
	fmt.Println("\n=== PART 1: IMAGE PROCESSING ===")

	// ============================================
	// STDLIB: IMAGE
	// ============================================
	fmt.Println("\n--- Package image ---")

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	red := color.RGBA{255, 0, 0, 255}
	for x := 0; x < 50; x++ {
		for y := 0; y < 50; y++ {
			img.Set(x, y, red)
		}
	}

	green := color.RGBA{0, 255, 0, 255}
	for x := 50; x < 100; x++ {
		for y := 0; y < 50; y++ {
			img.Set(x, y, green)
		}
	}

	blue := color.RGBA{0, 0, 255, 255}
	for x := 0; x < 50; x++ {
		for y := 50; y < 100; y++ {
			img.Set(x, y, blue)
		}
	}

	yellow := color.RGBA{255, 255, 0, 255}
	for x := 50; x < 100; x++ {
		for y := 50; y < 100; y++ {
			img.Set(x, y, yellow)
		}
	}

	fmt.Printf("Image created: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())

	r, g, b, a := img.At(25, 25).RGBA()
	fmt.Printf("Color at (25,25): R=%d G=%d B=%d A=%d\n", r>>8, g>>8, b>>8, a>>8)

	// ============================================
	// SAVE IMAGE
	// ============================================
	fmt.Println("\n--- Save Image ---")

	pngFile, err := os.Create("/tmp/test_image.png")
	if err == nil {
		png.Encode(pngFile, img)
		pngFile.Close()
		fmt.Println("Saved: /tmp/test_image.png")
	}

	jpgFile, err := os.Create("/tmp/test_image.jpg")
	if err == nil {
		jpeg.Encode(jpgFile, img, &jpeg.Options{Quality: 90})
		jpgFile.Close()
		fmt.Println("Saved: /tmp/test_image.jpg")
	}

	// ============================================
	// LOAD IMAGE
	// ============================================
	fmt.Println("\n--- Load Image ---")
	fmt.Println(`
REGISTER DECODERS:
import (
    "image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
)

LOAD FROM FILE:
file, err := os.Open("image.png")
if err != nil { log.Fatal(err) }
defer file.Close()

img, format, err := image.Decode(file)
img.Bounds().Size()

DIMENSIONS ONLY (no full decode):
config, format, err := image.DecodeConfig(file)`)

	// ============================================
	// DRAW
	// ============================================
	fmt.Println("\n--- Package draw ---")

	dst := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(50, 50, 150, 150), img, image.Point{}, draw.Over)
	fmt.Println("Composite image created")

	fmt.Println(`
DRAW OPERATIONS:
draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Src)   // Copy
draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Over)  // Alpha composite
draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Red}, image.Point{}, draw.Src)`)

	// ============================================
	// IMAGING LIBRARY
	// ============================================
	fmt.Println("\n--- disintegration/imaging ---")
	fmt.Println(`
import "github.com/disintegration/imaging"

src, err := imaging.Open("input.jpg")

RESIZE:
dst := imaging.Resize(src, 800, 0, imaging.Lanczos)
dst := imaging.Fill(src, 100, 100, imaging.Center, imaging.Lanczos)
dst := imaging.Fit(src, 800, 600, imaging.Lanczos)

ROTATE/FLIP:
dst := imaging.Rotate90(src)
dst := imaging.FlipH(src)

ADJUSTMENTS:
dst := imaging.AdjustBrightness(src, 20)
dst := imaging.AdjustContrast(src, 20)
dst := imaging.Grayscale(src)

FILTERS:
dst := imaging.Blur(src, 5.0)
dst := imaging.Sharpen(src, 3.0)

imaging.Save(dst, "output.jpg")`)

	// ============================================
	// BILD
	// ============================================
	fmt.Println("\n--- anthonynsimon/bild ---")
	fmt.Println(`
import "github.com/anthonynsimon/bild/effect"
import "github.com/anthonynsimon/bild/transform"

EFFECTS:
result := effect.Grayscale(img)
result := effect.Sepia(img)
result := effect.EdgeDetection(img, 1.0)
result := effect.Emboss(img)

TRANSFORMS:
result := transform.Resize(img, 800, 600, transform.Linear)
result := transform.Rotate(img, 45.0, nil)`)

	// ============================================
	// GOCV (OPENCV)
	// ============================================
	fmt.Println("\n--- GoCV (OpenCV) ---")
	fmt.Println(`
import "gocv.io/x/gocv"

img := gocv.IMRead("input.jpg", gocv.IMReadColor)
defer img.Close()

gray := gocv.NewMat()
gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

edges := gocv.NewMat()
gocv.Canny(gray, &edges, 50, 150)

FACE DETECTION:
classifier := gocv.NewCascadeClassifier()
classifier.Load("haarcascade_frontalface_default.xml")
rects := classifier.DetectMultiScale(gray)`)

	// ============================================
	// PART 2: AUDIO PROCESSING
	// ============================================
	fmt.Println("\n=== PART 2: AUDIO PROCESSING ===")

	// ============================================
	// AUDIO LIBRARIES
	// ============================================
	fmt.Println("\n--- Audio Libraries ---")
	fmt.Println(`
PLAYBACK/RECORDING:
- oto (github.com/hajimehoshi/oto)       Low-level audio I/O
- beep (github.com/faiface/beep)         High-level playback + effects
- portaudio (github.com/gordonklaus/portaudio)  Recording

FORMATS:
- go-mp3 (github.com/hajimehoshi/go-mp3)
- go-wav (github.com/youpy/go-wav)
- flac (github.com/mewkiz/flac)

DSP:
- go-dsp (github.com/mjibson/go-dsp)     FFT, filters`)

	// ============================================
	// OTO (LOW-LEVEL)
	// ============================================
	fmt.Println("\n--- Oto (Low-Level Audio) ---")
	fmt.Println(`
import "github.com/hajimehoshi/oto/v2"

ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
    SampleRate:   44100,
    ChannelCount: 2,
    Format:       oto.FormatSignedInt16LE,
})
<-ready

player := ctx.NewPlayer(audioReader)
player.Play()

SINE WAVE GENERATOR:
func generateSineWave(freq, duration float64, sampleRate int) []byte {
    numSamples := int(duration * float64(sampleRate))
    buf := make([]byte, numSamples*4)
    for i := 0; i < numSamples; i++ {
        t := float64(i) / float64(sampleRate)
        sample := int16(math.Sin(2*math.Pi*freq*t) * 32767 * 0.5)
        buf[i*4] = byte(sample)
        buf[i*4+1] = byte(sample >> 8)
        buf[i*4+2] = byte(sample)
        buf[i*4+3] = byte(sample >> 8)
    }
    return buf
}`)

	// ============================================
	// BEEP (HIGH-LEVEL)
	// ============================================
	fmt.Println("\n--- Beep (High-Level) ---")
	fmt.Println(`
import (
    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
    "github.com/faiface/beep/speaker"
    "github.com/faiface/beep/effects"
)

speaker.Init(sampleRate, sampleRate.N(time.Second/10))

f, _ := os.Open("music.mp3")
streamer, format, _ := mp3.Decode(f)
speaker.Play(streamer)

EFFECTS:
looped := beep.Loop(-1, streamer)
resampled := beep.Resample(4, format.SampleRate, 44100, streamer)
mixed := beep.Mix(sound1, sound2, sound3)
seq := beep.Seq(sound1, sound2)

VOLUME:
volume := &effects.Volume{Streamer: streamer, Base: 2, Volume: -1}`)

	// ============================================
	// WAV AND MP3
	// ============================================
	fmt.Println("\n--- WAV and MP3 ---")
	fmt.Println(`
WAV (github.com/youpy/go-wav):
reader := wav.NewReader(file)
format, _ := reader.Format()
samples, err := reader.ReadSamples()

MP3 (github.com/hajimehoshi/go-mp3):
decoder, _ := mp3.NewDecoder(file)
sampleRate := decoder.SampleRate()
buf := make([]byte, 8192)
n, err := decoder.Read(buf)`)

	// ============================================
	// DSP
	// ============================================
	fmt.Println("\n--- DSP (Digital Signal Processing) ---")
	fmt.Println(`
import "github.com/mjibson/go-dsp/fft"

FFT:
spectrum := fft.FFT(signal)
reconstructed := fft.IFFT(spectrum)
spectrum := fft.FFTReal(realSignal)

LOW-PASS FILTER:
func lowPassFilter(samples []float64, windowSize int) []float64 {
    result := make([]float64, len(samples))
    for i := range samples {
        sum, count := 0.0, 0
        for j := i - windowSize/2; j <= i+windowSize/2; j++ {
            if j >= 0 && j < len(samples) {
                sum += samples[j]
                count++
            }
        }
        result[i] = sum / float64(count)
    }
    return result
}`)

	// ============================================
	// PORTAUDIO
	// ============================================
	fmt.Println("\n--- PortAudio (Recording) ---")
	fmt.Println(`
import "github.com/gordonklaus/portaudio"

portaudio.Initialize()
defer portaudio.Terminate()

RECORD:
buf := make([]int16, 1024)
stream, _ := portaudio.OpenDefaultStream(1, 0, 44100, len(buf), &buf)
stream.Start()
stream.Read()

PLAYBACK:
stream, _ := portaudio.OpenDefaultStream(0, 2, 44100, len(buf), &buf)
stream.Start()
stream.Write()`)
}

/*
SUMMARY - IMAGES:

STDLIB:
image - types, interfaces
image/color - colors
image/draw - composition
image/png, image/jpeg, image/gif - codecs

CREATE:
img := image.NewRGBA(image.Rect(0, 0, w, h))
img.Set(x, y, color)

LOAD/SAVE:
image.Decode(reader) -> img, format, err
png.Encode(writer, img)
jpeg.Encode(writer, img, opts)

LIBRARIES:
disintegration/imaging - Resize, Crop, Rotate, Blur, Sharpen
anthonynsimon/bild - Effects: Sepia, EdgeDetection, Emboss
gocv.io/x/gocv - OpenCV: Computer vision, face detection

SUMMARY - AUDIO:

LIBRARIES:
oto - low-level I/O cross-platform
beep - high-level playback + effects
portaudio - recording + playback
go-dsp - FFT, filters

FORMATS: go-mp3, go-wav, flac, vorbis

EFFECTS (beep):
beep.Loop(), beep.Resample(), beep.Mix(), beep.Seq()
effects.Volume{}

DSP: FFT, IFFT, filters, normalization
*/
