// Package main - Chapter 088: Desktop Applications
// Go puede crear aplicaciones de escritorio cross-platform.
// Los principales frameworks son Fyne, Wails y Gio.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== DESKTOP APPLICATIONS EN GO ===")

	// ============================================
	// OPCIONES PRINCIPALES
	// ============================================
	fmt.Println("\n--- Opciones Principales ---")
	fmt.Println(`
FRAMEWORKS DE GUI:

1. FYNE (fyne.io)
   - Pure Go, no CGO requerido
   - Material Design look
   - Cross-platform (Windows, macOS, Linux, mobile)
   - Fácil de aprender

2. WAILS (wails.io)
   - Frontend: HTML/CSS/JavaScript (React, Vue, Svelte)
   - Backend: Go
   - Usa WebView nativo (no Electron)
   - Binarios pequeños

3. GIO (gioui.org)
   - Immediate mode GUI
   - Alto rendimiento
   - Curva de aprendizaje más alta
   - Mobile también

4. GTK (gotk3/gotk4)
   - Bindings para GTK
   - Requiere GTK instalado
   - Aspecto nativo en Linux`)
	// ============================================
	// FYNE
	// ============================================
	fmt.Println("\n--- Fyne ---")
	fmt.Println(`
INSTALACIÓN:
go get fyne.io/fyne/v2

HELLO WORLD:
package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/widget"
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("Hello")

    myWindow.SetContent(widget.NewLabel("Hello Fyne!"))
    myWindow.ShowAndRun()
}

WIDGETS BÁSICOS:
// Label
label := widget.NewLabel("Hello")

// Button
button := widget.NewButton("Click me", func() {
    fmt.Println("Clicked!")
})

// Entry (input)
entry := widget.NewEntry()
entry.SetPlaceHolder("Enter text...")

password := widget.NewPasswordEntry()

// Check
check := widget.NewCheck("Enable feature", func(checked bool) {
    fmt.Println("Checked:", checked)
})

// Radio
radio := widget.NewRadioGroup([]string{"Option 1", "Option 2"}, func(selected string) {
    fmt.Println("Selected:", selected)
})

// Select (dropdown)
sel := widget.NewSelect([]string{"A", "B", "C"}, func(selected string) {
    fmt.Println("Selected:", selected)
})

// Slider
slider := widget.NewSlider(0, 100)
slider.OnChanged = func(value float64) {
    fmt.Println("Value:", value)
}

LAYOUTS:
import "fyne.io/fyne/v2/container"

// VBox - vertical
vbox := container.NewVBox(
    widget.NewLabel("Top"),
    widget.NewButton("Middle", nil),
    widget.NewLabel("Bottom"),
)

// HBox - horizontal
hbox := container.NewHBox(label, button, entry)

// Grid
grid := container.NewGridWithColumns(3, items...)

// Border
border := container.NewBorder(top, bottom, left, right, center)

// Tabs
tabs := container.NewAppTabs(
    container.NewTabItem("Tab 1", content1),
    container.NewTabItem("Tab 2", content2),
)

// Split
split := container.NewHSplit(left, right)
split.SetOffset(0.3)

DIÁLOGOS:
import "fyne.io/fyne/v2/dialog"

// Info
dialog.ShowInformation("Title", "Message", window)

// Confirm
dialog.ShowConfirm("Title", "Are you sure?", func(ok bool) {
    if ok {
        // confirmed
    }
}, window)

// File open
dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
    if err != nil || reader == nil {
        return
    }
    defer reader.Close()
    // Read file
}, window)

// File save
dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
    // Write file
}, window)

DATA BINDING:
import "fyne.io/fyne/v2/data/binding"

// Bind string
data := binding.NewString()
data.Set("Initial value")

label := widget.NewLabelWithData(data)
entry := widget.NewEntryWithData(data)

// Cuando entry cambia, label se actualiza automáticamente

// Lista con binding
list := binding.NewStringList()
list.Append("Item 1")
listWidget := widget.NewListWithData(list, ...)

BUNDLE RESOURCES:
//go:embed icon.png
var iconData []byte

icon := fyne.NewStaticResource("icon.png", iconData)
window.SetIcon(icon)

BUILD:
# Desarrollo
go run .

# Producción (con fyne tool)
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os darwin -icon icon.png
fyne package -os windows -icon icon.png
fyne package -os linux -icon icon.png`)
	// ============================================
	// WAILS
	// ============================================
	fmt.Println("\n--- Wails ---")
	fmt.Println(`
INSTALACIÓN:
go install github.com/wailsapp/wails/v2/cmd/wails@latest

CREAR PROYECTO:
wails init -n myapp -t vue
# o -t react, -t svelte, -t vanilla

ESTRUCTURA:
myapp/
├── build/           # Build output
├── frontend/        # Web frontend
│   ├── src/
│   └── package.json
├── app.go           # Backend Go
├── main.go
└── wails.json

// main.go
package main

import (
    "embed"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:  "My App",
        Width:  1024,
        Height: 768,
        Assets: assets,
        Bind: []interface{}{
            app,
        },
    })
    if err != nil {
        println("Error:", err)
    }
}

// app.go
package main

import "context"

type App struct {
    ctx context.Context
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
}

// Exportado al frontend
func (a *App) Greet(name string) string {
    return "Hello " + name + "!"
}

func (a *App) GetUsers() ([]User, error) {
    // Fetch from database
    return users, nil
}

// En frontend (JavaScript/TypeScript):
import { Greet, GetUsers } from '../wailsjs/go/main/App';

async function greet() {
    const result = await Greet("World");
    console.log(result);  // "Hello World!"
}

EVENTOS:
// Go -> Frontend
runtime.EventsEmit(a.ctx, "notification", "New message!")

// Frontend -> Go
runtime.EventsOn("user-action", func(optionalData ...interface{}) {
    fmt.Println("User did something")
})

// En frontend:
EventsOn("notification", (message) => {
    console.log(message);
});

EventsEmit("user-action", data);

BUILD:
wails dev      # Development con hot reload
wails build    # Producción

# Cross-compile
wails build -platform darwin/arm64
wails build -platform windows/amd64
wails build -platform linux/amd64`)
	// ============================================
	// GIO
	// ============================================
	fmt.Println("\n--- Gio ---")
	fmt.Println(`
INSTALACIÓN:
go get gioui.org

HELLO WORLD:
package main

import (
    "gioui.org/app"
    "gioui.org/layout"
    "gioui.org/op"
    "gioui.org/widget/material"
    "gioui.org/font/gofont"
)

func main() {
    go func() {
        w := new(app.Window)
        if err := loop(w); err != nil {
            log.Fatal(err)
        }
        os.Exit(0)
    }()
    app.Main()
}

func loop(w *app.Window) error {
    th := material.NewTheme()
    th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
    var ops op.Ops

    for {
        switch e := w.Event().(type) {
        case app.DestroyEvent:
            return e.Err
        case app.FrameEvent:
            gtx := app.NewContext(&ops, e)
            layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
                return material.H1(th, "Hello, Gio!").Layout(gtx)
            })
            e.Frame(gtx.Ops)
        }
    }
}

WIDGETS:
// Button
var button widget.Clickable

btn := material.Button(th, &button, "Click me")
btn.Layout(gtx)

if button.Clicked(gtx) {
    fmt.Println("Clicked!")
}

// Editor (text input)
var editor widget.Editor

ed := material.Editor(th, &editor, "Hint")
ed.Layout(gtx)

text := editor.Text()

// List
var list widget.List

material.List(th, &list).Layout(gtx, len(items), func(gtx C, i int) D {
    return material.Body1(th, items[i]).Layout(gtx)
})

CARACTERÍSTICAS:
- Immediate mode (redibuja cada frame)
- Alto rendimiento
- GPU accelerated
- Touch support
- Más control sobre rendering
- Curva de aprendizaje más pronunciada`)
	// ============================================
	// SYSTRAY
	// ============================================
	fmt.Println("\n--- System Tray ---")
	fmt.Println(`
github.com/getlantern/systray

package main

import (
    "github.com/getlantern/systray"
)

func main() {
    systray.Run(onReady, onExit)
}

func onReady() {
    systray.SetIcon(iconData)
    systray.SetTitle("My App")
    systray.SetTooltip("My App is running")

    mShow := systray.AddMenuItem("Show", "Show window")
    mHide := systray.AddMenuItem("Hide", "Hide window")
    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "Quit the app")

    go func() {
        for {
            select {
            case <-mShow.ClickedCh:
                // Show window
            case <-mHide.ClickedCh:
                // Hide window
            case <-mQuit.ClickedCh:
                systray.Quit()
            }
        }
    }()
}

func onExit() {
    // Cleanup
}`)
	// ============================================
	// COMPARACIÓN
	// ============================================
	fmt.Println("\n--- Comparación ---")
	fmt.Println(`
| Feature           | Fyne      | Wails     | Gio       |
|-------------------|-----------|-----------|-----------|
| Lenguaje UI       | Go        | Web       | Go        |
| CGO requerido     | No        | No        | Sí(GPU)   |
| Tamaño binario    | ~15MB     | ~8MB      | ~10MB     |
| Look & Feel       | Material  | Nativo*   | Custom    |
| Curva aprendizaje | Fácil     | Medio     | Difícil   |
| Performance       | Bueno     | Muy bueno | Excelente |
| Mobile            | Sí        | No        | Sí        |
| Hot reload        | No        | Sí        | No        |

* Wails usa WebView nativo, así que el look depende del CSS

RECOMENDACIONES:

FYNE:
- Principiantes en desktop Go
- Apps internas/utilities
- Cuando necesitas mobile también

WAILS:
- Desarrolladores web
- UI complejas con frameworks JS modernos
- Cuando el aspecto visual es crítico

GIO:
- Alto rendimiento requerido
- Control total sobre rendering
- Experiencia en immediate mode GUIs`)
	// ============================================
	// DISTRIBUCIÓN
	// ============================================
	fmt.Println("\n--- Distribución ---")
	fmt.Println(`
WINDOWS:
- .exe directo
- Installer con NSIS o WiX
- MSIX para Microsoft Store

MACOS:
- .app bundle
- DMG para distribución
- Notarización para Gatekeeper
- PKG para instalación

LINUX:
- AppImage (portable)
- Flatpak
- Snap
- .deb / .rpm

HERRAMIENTAS:
- fyne package (para Fyne)
- wails build (para Wails)
- go-selfupdate (auto-actualizaciones)
- goreleaser (builds automatizados)

EJEMPLO GORELEASER:
# .goreleaser.yaml
builds:
  - id: myapp
    main: ./cmd/myapp
    goos: [darwin, linux, windows]
    goarch: [amd64, arm64]

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

nfpms:
  - formats: [deb, rpm]
    maintainer: Your Name <email@example.com>`)}

/*
RESUMEN DE DESKTOP APPS:

FRAMEWORKS:
- Fyne: Pure Go, fácil, cross-platform
- Wails: Go backend + Web frontend
- Gio: Immediate mode, alto rendimiento

FYNE:
app.New()
window.SetContent(widget)
window.ShowAndRun()

WAILS:
wails init -n app -t react
Bind Go methods -> Call from JS
wails build

GIO:
Immediate mode rendering
Layout system
Event handling manual

SYSTEM TRAY:
systray.Run(onReady, onExit)
systray.AddMenuItem()

DISTRIBUCIÓN:
- Windows: .exe, installer
- macOS: .app bundle, DMG
- Linux: AppImage, Flatpak, Snap

RECOMENDACIONES:
1. Empezar con Fyne para aprender
2. Wails para UIs web-like
3. Gio para rendimiento crítico
*/

/* SUMMARY - DESKTOP APPLICATIONS IN GO:

TOPIC: Creating cross-platform desktop applications with Go GUI frameworks
- Three main frameworks: Fyne (pure Go), Wails (Go + Web), Gio (immediate mode)
- Fyne uses Material Design, no CGO required, cross-platform including mobile
- Fyne widgets: Label, Button, Entry, Check, Radio, Select, Slider
- Fyne layouts: VBox, HBox, Grid, Border, Tabs, Split containers
- Fyne dialogs: ShowInformation, ShowConfirm, ShowFileOpen, ShowFileSave
- Data binding in Fyne with binding.NewString() for reactive UI updates
- fyne package command for building platform-specific bundles
- Wails combines Go backend with HTML/CSS/JavaScript frontend using native WebView
- wails init creates project with React, Vue, Svelte, or vanilla JavaScript
- Wails binds Go methods automatically callable from JavaScript
- EventsEmit and EventsOn for bidirectional communication in Wails
- wails dev for hot reload development, wails build for production
- Gio uses immediate mode rendering with high performance and GPU acceleration
- Gio layout system with layout.Context and custom rendering control
- systray package for system tray icons with menu items
- Distribution: Windows (.exe, NSIS), macOS (.app, DMG, notarization), Linux (AppImage, Flatpak, Snap)
- goreleaser for automated multi-platform builds
- Binary sizes: Wails ~8MB, Gio ~10MB, Fyne ~15MB
- Choose Fyne for simplicity, Wails for web-like UI, Gio for performance
*/
