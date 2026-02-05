// Package main - Capítulo 43: Procesamiento de Audio
// Go puede trabajar con audio para playback, grabación,
// y procesamiento de señales.
package main

import (
	"os"
	"fmt"
)

func main() {
	fmt.Println("=== PROCESAMIENTO DE AUDIO EN GO ===")

	// ============================================
	// BIBLIOTECAS PRINCIPALES
	// ============================================
	fmt.Println("\n--- Bibliotecas Principales ---")
	fmt.Println(`
PLAYBACK/RECORDING:
- oto (github.com/hajimehoshi/oto)
  Low-level audio I/O, cross-platform

- beep (github.com/faiface/beep)
  High-level audio playback y efectos

- portaudio (github.com/gordonklaus/portaudio)
  Bindings para PortAudio

FORMATOS:
- go-mp3 (github.com/hajimehoshi/go-mp3)
- go-wav (github.com/youpy/go-wav)
- flac (github.com/mewkiz/flac)
- vorbis (github.com/jfreymuth/vorbis)

PROCESAMIENTO:
- go-dsp (github.com/mjibson/go-dsp)
  FFT, filtros, análisis`)
	// ============================================
	// OTO (LOW-LEVEL)
	// ============================================
	fmt.Println("\n--- Oto (Low-Level Audio) ---")
	fmt.Println(`
import "github.com/hajimehoshi/oto/v2"

// Crear contexto de audio
ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
    SampleRate:   44100,
    ChannelCount: 2,
    Format:       oto.FormatSignedInt16LE,
})
<-ready  // Esperar a que esté listo

// Crear player
player := ctx.NewPlayer(audioReader)

// Reproducir
player.Play()

// Esperar a que termine
for player.IsPlaying() {
    time.Sleep(time.Millisecond * 100)
}

// Generar tono simple
func generateSineWave(freq, duration float64, sampleRate int) []byte {
    numSamples := int(duration * float64(sampleRate))
    buf := make([]byte, numSamples*4) // 2 channels * 2 bytes

    for i := 0; i < numSamples; i++ {
        t := float64(i) / float64(sampleRate)
        sample := int16(math.Sin(2*math.Pi*freq*t) * 32767 * 0.5)

        // Left channel
        buf[i*4] = byte(sample)
        buf[i*4+1] = byte(sample >> 8)
        // Right channel
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
    "github.com/faiface/beep/wav"
    "github.com/faiface/beep/speaker"
    "github.com/faiface/beep/effects"
)

// Inicializar speaker
speaker.Init(sampleRate, sampleRate.N(time.Second/10))

// Cargar MP3
f, _ := os.Open("music.mp3")
streamer, format, _ := mp3.Decode(f)
defer streamer.Close()

// Reproducir
speaker.Play(streamer)

// Esperar a que termine
select {}

// Control de volumen
ctrl := &beep.Ctrl{Streamer: streamer}
volume := &effects.Volume{
    Streamer: ctrl,
    Base:     2,
    Volume:   -1,  // -1 = 50% volume
    Silent:   false,
}
speaker.Play(volume)

// Efectos
// Loop
looped := beep.Loop(-1, streamer)  // -1 = infinito

// Resample (cambiar sample rate)
resampled := beep.Resample(4, format.SampleRate, 44100, streamer)

// Secuencia
seq := beep.Seq(
    sound1,
    beep.Callback(func() { fmt.Println("Sound 1 done") }),
    sound2,
)

// Mezclar (reproducir simultáneamente)
mixed := beep.Mix(sound1, sound2, sound3)

// Pausa
ctrl.Paused = true
ctrl.Paused = false`)
	// ============================================
	// WAV FILES
	// ============================================
	fmt.Println("\n--- Archivos WAV ---")
	os.Stdout.WriteString(`
import "github.com/youpy/go-wav"

// Leer WAV
file, _ := os.Open("audio.wav")
reader := wav.NewReader(file)

format, _ := reader.Format()
fmt.Printf("Sample Rate: %d, Channels: %d, Bits: %d\n",
    format.SampleRate, format.NumChannels, format.BitsPerSample)

// Leer samples
for {
    samples, err := reader.ReadSamples()
    if err == io.EOF {
        break
    }
    for _, sample := range samples {
        // sample.Values[0] = left channel
        // sample.Values[1] = right channel (si stereo)
    }
}

// Escribir WAV
import "github.com/go-audio/wav"

outFile, _ := os.Create("output.wav")
encoder := wav.NewEncoder(outFile, 44100, 16, 2, 1)
defer encoder.Close()

// Escribir samples
buf := &audio.IntBuffer{
    Data:           samples,
    Format:         &audio.Format{SampleRate: 44100, NumChannels: 2},
    SourceBitDepth: 16,
}
encoder.Write(buf)
`)

	// ============================================
	// MP3
	// ============================================
	fmt.Println("\n--- MP3 ---")
	os.Stdout.WriteString(`
import "github.com/hajimehoshi/go-mp3"

// Decodificar MP3
file, _ := os.Open("music.mp3")
decoder, _ := mp3.NewDecoder(file)

// Info
sampleRate := decoder.SampleRate()
length := decoder.Length()
fmt.Printf("Sample Rate: %d, Length: %d bytes\n", sampleRate, length)

// Leer PCM data
buf := make([]byte, 8192)
for {
    n, err := decoder.Read(buf)
    if err == io.EOF {
        break
    }
    // Procesar buf[:n]
}

// Con beep
import "github.com/faiface/beep/mp3"

file, _ := os.Open("music.mp3")
streamer, format, _ := mp3.Decode(file)
// format.SampleRate, format.NumChannels
`)

	// ============================================
	// DSP
	// ============================================
	fmt.Println("\n--- DSP (Digital Signal Processing) ---")
	os.Stdout.WriteString(`
import "github.com/mjibson/go-dsp/fft"

// FFT (Fast Fourier Transform)
signal := []complex128{ /* samples */ }
spectrum := fft.FFT(signal)

// Inverse FFT
reconstructed := fft.IFFT(spectrum)

// FFT de datos reales
realSignal := []float64{ /* samples */ }
spectrum := fft.FFTReal(realSignal)

// Magnitud del espectro
for i, c := range spectrum {
    magnitude := cmplx.Abs(c)
    frequency := float64(i) * sampleRate / float64(len(spectrum))
    fmt.Printf("Freq: %.1f Hz, Mag: %.2f\n", frequency, magnitude)
}

FILTROS:
// Low-pass filter simple (moving average)
func lowPassFilter(samples []float64, windowSize int) []float64 {
    result := make([]float64, len(samples))
    for i := range samples {
        sum := 0.0
        count := 0
        for j := i - windowSize/2; j <= i+windowSize/2; j++ {
            if j >= 0 && j < len(samples) {
                sum += samples[j]
                count++
            }
        }
        result[i] = sum / float64(count)
    }
    return result
}

// Normalización
func normalize(samples []float64) []float64 {
    maxVal := 0.0
    for _, s := range samples {
        if abs := math.Abs(s); abs > maxVal {
            maxVal = abs
        }
    }
    result := make([]float64, len(samples))
    for i, s := range samples {
        result[i] = s / maxVal
    }
    return result
}
`)

	// ============================================
	// PORTAUDIO
	// ============================================
	fmt.Println("\n--- PortAudio (Recording) ---")
	os.Stdout.WriteString(`
import "github.com/gordonklaus/portaudio"

portaudio.Initialize()
defer portaudio.Terminate()

// Listar dispositivos
devices, _ := portaudio.Devices()
for _, d := range devices {
    fmt.Printf("%s: in=%d, out=%d\n", d.Name, d.MaxInputChannels, d.MaxOutputChannels)
}

// Grabar
buf := make([]int16, 1024)
stream, _ := portaudio.OpenDefaultStream(1, 0, 44100, len(buf), &buf)
defer stream.Close()

stream.Start()
defer stream.Stop()

for {
    stream.Read()
    // Procesar buf
}

// Reproducir
stream, _ := portaudio.OpenDefaultStream(0, 2, 44100, len(buf), &buf)
stream.Start()

for {
    // Llenar buf con audio
    stream.Write()
}

// Dispositivo específico
params := portaudio.LowLatencyParameters(inputDevice, outputDevice)
params.Input.Channels = 1
params.Output.Channels = 2
params.SampleRate = 44100
params.FramesPerBuffer = 1024

stream, _ := portaudio.OpenStream(params, callback)
`)

	// ============================================
	// EJEMPLO: GENERADOR DE TONOS
	// ============================================
	fmt.Println("\n--- Ejemplo: Generador de Tonos ---")
	fmt.Println(`
package main

import (
    "math"
    "time"

    "github.com/hajimehoshi/oto/v2"
)

type ToneGenerator struct {
    sampleRate int
    frequency  float64
    phase      float64
}

func (t *ToneGenerator) Read(buf []byte) (int, error) {
    for i := 0; i < len(buf)/4; i++ {
        sample := int16(math.Sin(t.phase) * 32767 * 0.3)
        t.phase += 2 * math.Pi * t.frequency / float64(t.sampleRate)

        // Left
        buf[i*4] = byte(sample)
        buf[i*4+1] = byte(sample >> 8)
        // Right
        buf[i*4+2] = byte(sample)
        buf[i*4+3] = byte(sample >> 8)
    }
    return len(buf), nil
}

func main() {
    ctx, ready, _ := oto.NewContext(&oto.NewContextOptions{
        SampleRate:   44100,
        ChannelCount: 2,
        Format:       oto.FormatSignedInt16LE,
    })
    <-ready

    gen := &ToneGenerator{
        sampleRate: 44100,
        frequency:  440, // A4
    }

    player := ctx.NewPlayer(gen)
    player.Play()

    // Reproducir por 2 segundos
    time.Sleep(2 * time.Second)
}`)}

/*
RESUMEN DE AUDIO:

BIBLIOTECAS:
- oto: low-level I/O cross-platform
- beep: high-level playback + efectos
- portaudio: recording + playback

FORMATOS:
- go-mp3: decodificar MP3
- go-wav: leer/escribir WAV
- flac, vorbis: otros formatos

DSP:
- go-dsp: FFT, filtros
- Análisis de frecuencias

PLAYBACK (beep):
speaker.Init(sampleRate, bufferSize)
speaker.Play(streamer)

EFECTOS:
beep.Loop() - repetir
beep.Resample() - cambiar sample rate
effects.Volume{} - volumen
beep.Seq() - secuencia
beep.Mix() - mezclar

RECORDING (portaudio):
portaudio.OpenDefaultStream()
stream.Read() / stream.Write()

PROCESAMIENTO:
- FFT para análisis de frecuencias
- Filtros (low-pass, high-pass)
- Normalización
- Mezcla

CASOS DE USO:
- Media players
- Generación de tonos
- Análisis de audio
- Voice recording
- Efectos de sonido
*/
