// Package main - Capítulo 42: Procesamiento de Imágenes
// Go tiene soporte nativo para imágenes en la stdlib
// y bibliotecas adicionales para manipulación avanzada.
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
	fmt.Println("=== PROCESAMIENTO DE IMÁGENES EN GO ===")

	// ============================================
	// STDLIB: IMAGE
	// ============================================
	fmt.Println("\n--- Paquete image ---")

	// Crear imagen RGBA
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Dibujar píxeles
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

	fmt.Printf("Imagen creada: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())

	// Leer pixel
	r, g, b, a := img.At(25, 25).RGBA()
	fmt.Printf("Color en (25,25): R=%d G=%d B=%d A=%d\n", r>>8, g>>8, b>>8, a>>8)

	// ============================================
	// GUARDAR IMAGEN
	// ============================================
	fmt.Println("\n--- Guardar Imagen ---")

	// Guardar como PNG
	pngFile, err := os.Create("/tmp/test_image.png")
	if err == nil {
		png.Encode(pngFile, img)
		pngFile.Close()
		fmt.Println("Guardado: /tmp/test_image.png")
	}

	// Guardar como JPEG
	jpgFile, err := os.Create("/tmp/test_image.jpg")
	if err == nil {
		jpeg.Encode(jpgFile, img, &jpeg.Options{Quality: 90})
		jpgFile.Close()
		fmt.Println("Guardado: /tmp/test_image.jpg")
	}

	// ============================================
	// CARGAR IMAGEN
	// ============================================
	fmt.Println("\n--- Cargar Imagen ---")
	os.Stdout.WriteString(`
// Registrar decoders
import (
    "image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
)

// Cargar desde archivo
file, err := os.Open("image.png")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Decodificar (detecta formato automáticamente)
img, format, err := image.Decode(file)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Formato: %s, Tamaño: %v\n", format, img.Bounds().Size())

// Solo obtener dimensiones (sin decodificar todo)
file.Seek(0, 0)  // Regresar al inicio
config, format, err := image.DecodeConfig(file)
fmt.Printf("Dimensiones: %dx%d\n", config.Width, config.Height)
`)

	// ============================================
	// DRAW
	// ============================================
	fmt.Println("\n--- Paquete draw ---")

	// Crear imagen de destino
	dst := image.NewRGBA(image.Rect(0, 0, 200, 200))

	// Llenar con color de fondo
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Dibujar otra imagen encima
	draw.Draw(dst, image.Rect(50, 50, 150, 150), img, image.Point{}, draw.Over)

	fmt.Println("Imagen compuesta creada")

	fmt.Println(`
OPERACIONES CON DRAW:

// Copiar imagen
draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Src)

// Dibujar con transparencia (composición)
draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Over)

// Llenar con color sólido
draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Red}, image.Point{}, draw.Src)

// Recortar y copiar
srcRect := image.Rect(10, 10, 50, 50)
draw.Draw(dst, dstRect, src, srcRect.Min, draw.Src)`)
	// ============================================
	// IMAGING LIBRARY
	// ============================================
	fmt.Println("\n--- disintegration/imaging ---")
	fmt.Println(`
import "github.com/disintegration/imaging"

// Abrir imagen
src, err := imaging.Open("input.jpg")

// Redimensionar
dst := imaging.Resize(src, 800, 0, imaging.Lanczos)  // Mantener aspecto
dst := imaging.Fill(src, 100, 100, imaging.Center, imaging.Lanczos)  // Crop
dst := imaging.Fit(src, 800, 600, imaging.Lanczos)   // Fit en rectángulo

// Rotar y voltear
dst := imaging.Rotate90(src)
dst := imaging.Rotate180(src)
dst := imaging.Rotate270(src)
dst := imaging.FlipH(src)   // Horizontal
dst := imaging.FlipV(src)   // Vertical

// Recortar
dst := imaging.Crop(src, image.Rect(0, 0, 100, 100))
dst := imaging.CropCenter(src, 200, 200)

// Ajustes de color
dst := imaging.AdjustBrightness(src, 20)     // -100 a 100
dst := imaging.AdjustContrast(src, 20)       // -100 a 100
dst := imaging.AdjustSaturation(src, 30)     // -100 a 100
dst := imaging.AdjustGamma(src, 1.2)
dst := imaging.Grayscale(src)
dst := imaging.Invert(src)

// Filtros
dst := imaging.Blur(src, 5.0)
dst := imaging.Sharpen(src, 3.0)

// Guardar
imaging.Save(dst, "output.jpg")
imaging.Save(dst, "output.png")`)
	// ============================================
	// BILD
	// ============================================
	fmt.Println("\n--- anthonynsimon/bild ---")
	fmt.Println(`
import "github.com/anthonynsimon/bild/imgio"
import "github.com/anthonynsimon/bild/transform"
import "github.com/anthonynsimon/bild/effect"
import "github.com/anthonynsimon/bild/adjust"
import "github.com/anthonynsimon/bild/blur"

// Cargar
img, err := imgio.Open("input.png")

// Transformaciones
result := transform.Resize(img, 800, 600, transform.Linear)
result := transform.Rotate(img, 45.0, nil)
result := transform.FlipH(img)
result := transform.Crop(img, image.Rect(0, 0, 100, 100))

// Efectos
result := effect.Grayscale(img)
result := effect.Sepia(img)
result := effect.EdgeDetection(img, 1.0)
result := effect.Emboss(img)
result := effect.Invert(img)
result := effect.Dilate(img, 3)
result := effect.Erode(img, 3)

// Ajustes
result := adjust.Brightness(img, 0.2)   // -1.0 a 1.0
result := adjust.Contrast(img, 0.2)
result := adjust.Saturation(img, 0.3)
result := adjust.Hue(img, 45)

// Blur
result := blur.Gaussian(img, 5.0)
result := blur.Box(img, 5.0)

// Guardar
imgio.Save("output.png", result, imgio.PNGEncoder())
imgio.Save("output.jpg", result, imgio.JPEGEncoder(90))`)
	// ============================================
	// GOCV (OPENCV)
	// ============================================
	fmt.Println("\n--- GoCV (OpenCV) ---")
	os.Stdout.WriteString(`
import "gocv.io/x/gocv"

// Leer imagen
img := gocv.IMRead("input.jpg", gocv.IMReadColor)
defer img.Close()

// Info
fmt.Printf("Size: %dx%d, Channels: %d\n", img.Cols(), img.Rows(), img.Channels())

// Convertir a grayscale
gray := gocv.NewMat()
defer gray.Close()
gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

// Blur
blurred := gocv.NewMat()
defer blurred.Close()
gocv.GaussianBlur(img, &blurred, image.Pt(5, 5), 0, 0, gocv.BorderDefault)

// Edge detection
edges := gocv.NewMat()
defer edges.Close()
gocv.Canny(gray, &edges, 50, 150)

// Redimensionar
resized := gocv.NewMat()
defer resized.Close()
gocv.Resize(img, &resized, image.Pt(800, 600), 0, 0, gocv.InterpolationLinear)

// Detección de rostros
classifier := gocv.NewCascadeClassifier()
classifier.Load("haarcascade_frontalface_default.xml")
defer classifier.Close()

rects := classifier.DetectMultiScale(gray)
for _, rect := range rects {
    gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 0}, 2)
}

// Video capture
webcam, _ := gocv.OpenVideoCapture(0)
defer webcam.Close()

window := gocv.NewWindow("Webcam")
defer window.Close()

frame := gocv.NewMat()
defer frame.Close()

for {
    webcam.Read(&frame)
    if frame.Empty() {
        continue
    }
    window.IMShow(frame)
    if window.WaitKey(1) >= 0 {
        break
    }
}

// Guardar
gocv.IMWrite("output.jpg", img)
`)

	// ============================================
	// EJEMPLO PRÁCTICO
	// ============================================
	fmt.Println("\n--- Ejemplo: Thumbnail Generator ---")
	os.Stdout.WriteString(`
package main

import (
    "fmt"
    "path/filepath"
    "github.com/disintegration/imaging"
)

func generateThumbnails(inputPath string, sizes []int) error {
    src, err := imaging.Open(inputPath)
    if err != nil {
        return err
    }

    dir := filepath.Dir(inputPath)
    base := filepath.Base(inputPath)
    ext := filepath.Ext(base)
    name := base[:len(base)-len(ext)]

    for _, size := range sizes {
        thumb := imaging.Fill(src, size, size, imaging.Center, imaging.Lanczos)

        outputPath := filepath.Join(dir, fmt.Sprintf("%s_%dx%d%s", name, size, size, ext))

        if err := imaging.Save(thumb, outputPath); err != nil {
            return err
        }

        fmt.Printf("Generated: %s\n", outputPath)
    }

    return nil
}

func main() {
    sizes := []int{64, 128, 256, 512}
    if err := generateThumbnails("photo.jpg", sizes); err != nil {
        log.Fatal(err)
    }
}
`)
}

/*
RESUMEN DE IMÁGENES:

STDLIB:
image - tipos básicos, interfaces
image/color - colores
image/draw - composición
image/png, image/jpeg, image/gif - codecs

CREAR IMAGEN:
img := image.NewRGBA(image.Rect(0, 0, w, h))
img.Set(x, y, color)
img.At(x, y) -> color

CARGAR/GUARDAR:
image.Decode(reader) -> img, format, err
png.Encode(writer, img)
jpeg.Encode(writer, img, &Options{Quality: 90})

LIBRERÍAS:
disintegration/imaging
- Resize, Crop, Rotate, Flip
- Ajustes: Brightness, Contrast, Saturation
- Filtros: Blur, Sharpen, Grayscale

anthonynsimon/bild
- Efectos: Sepia, EdgeDetection, Emboss
- Blur: Gaussian, Box
- Ajustes de color

gocv.io/x/gocv (OpenCV)
- Computer vision avanzado
- Detección de objetos/rostros
- Video processing

CASOS DE USO:
- Thumbnails
- Watermarks
- Filtros de imagen
- Detección de objetos
- OCR preprocessing
*/
