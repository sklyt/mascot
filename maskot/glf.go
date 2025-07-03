package mascot

import (
	"embed"
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

//go:embed shaders/*
var shaders embed.FS

// DesktopMascot displays a transparent GLFW window with a bottom-left animated sprite.
type DesktopMascot struct {
	window     *glfw.Window
	textures   []uint32
	frameDur   time.Duration
	frames     int
	shaderProg uint32
	vao, vbo   uint32
	OnClick    func() // Add click handler
}

var (
	MaskotInstance *DesktopMascot
	once           sync.Once
)

func GetMaskot(framePaths []string, frameRate float64) *DesktopMascot {
	once.Do(func() {
		m, err := newDesktopMascot(framePaths, frameRate)
		if err != nil {
			panic(err)
		}
		MaskotInstance = m
	})
	return MaskotInstance
}

// NewDesktopMascot initializes the window, compiles shaders, loads textures, and prepares rendering.
func newDesktopMascot(framePaths []string, frameRate float64) (*DesktopMascot, error) {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("failed to init GLFW: %w", err)
	}
	// Configure for OpenGL 3.3 core, transparent, borderless
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	glfw.WindowHint(glfw.Floating, glfw.True) // Always on top
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	win, err := glfw.CreateWindow(200, 250, "Mascot", nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	// Position bottom-left
	mon := glfw.GetPrimaryMonitor()
	x, y := mon.GetPos()
	mode := mon.GetVideoMode()
	win.SetPos(x, y+mode.Height-300)
	win.MakeContextCurrent()
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to init GL: %w", err)
	}
	gl.ClearColor(0, 0, 0, 0)

	// Load and compile shaders

	vertBytes, err := shaders.ReadFile("shaders/quad.vert")
	if err != nil {
		return nil, fmt.Errorf("vertex shader read error: %w", err)
	}
	fragBytes, err := shaders.ReadFile("shaders/quad.frag")
	if err != nil {
		return nil, fmt.Errorf("fragment shader read error: %w", err)
	}
	program, err := newProgram(string(vertBytes), string(fragBytes))
	if err != nil {
		return nil, fmt.Errorf("shader program error: %w", err)
	}
	// Bind sampler uniform once
	gl.UseProgram(program)
	samplerLoc := gl.GetUniformLocation(program, gl.Str("uTex\x00"))
	if samplerLoc < 0 {
		return nil, fmt.Errorf("uTex uniform not found")
	}
	gl.Uniform1i(samplerLoc, 0) // texture unit 0

	// Setup quad geometry
	var vao, vbo uint32
	setupQuad(&vao, &vbo)

	// Load sprite textures
	textures, err := loadTextures(framePaths)
	if err != nil {
		return nil, fmt.Errorf("texture load error: %w", err)
	}

	return &DesktopMascot{
		window:     win,
		textures:   textures,
		frameDur:   time.Duration(1e9 / frameRate),
		frames:     len(textures),
		shaderProg: program,
		vao:        vao,
		vbo:        vbo,
	}, nil
}
func (m *DesktopMascot) clean() {
	defer gl.DeleteProgram(m.shaderProg)
	defer gl.DeleteVertexArrays(1, &m.vao)
	defer gl.DeleteBuffers(1, &m.vbo)
	for _, tex := range m.textures {
		gl.DeleteTextures(1, &tex)
	}
}

func (m *DesktopMascot) Close_() {

	m.clean()
	m.window.Destroy()

}

// Run enters the render loop; check NewDesktopMascot error before calling.
func (m *DesktopMascot) Run() {

	defer m.clean()
	if m == nil {
		panic("DesktopMascot is nil; ensure NewDesktopMascot did not return an error")
	}
	var dragStart struct {
		x, y   float64
		active bool
	}
	m.window.SetCursorPosCallback(func(w *glfw.Window, x, y float64) {
		if dragStart.active {
			winX, winY := w.GetPos()
			w.SetPos(winX+int(x-dragStart.x), winY+int(y-dragStart.y))
		}
	})
	start := time.Now()
	m.window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button == glfw.MouseButtonLeft {

			// Existing click handler
			if action == glfw.Press && m.OnClick != nil {
				m.OnClick()
			}
		}

		if button == glfw.MouseButtonRight {
			if action == glfw.Press {
				dragStart.active = true
				dragStart.x, dragStart.y = w.GetCursorPos()
			} else {
				dragStart.active = false
			}
		}
	})
	m.window.Show()
	for !m.window.ShouldClose() {
		delta := time.Since(start)
		frame := int(delta/m.frameDur) % m.frames

		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.UseProgram(m.shaderProg)
		gl.BindVertexArray(m.vao)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, m.textures[frame])
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

		m.window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()

}

func setupQuad(vao, vbo *uint32) {
	vertices := []float32{
		-1, -1, 0, 1, // Bottom-left
		1, -1, 1, 1, // Bottom-right
		1, 1, 1, 0, // Top-right
		-1, 1, 0, 0, // Top-left
	}
	gl.GenVertexArrays(1, vao)
	gl.GenBuffers(1, vbo)
	gl.BindVertexArray(*vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, *vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	// pos
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	// uv
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
}

func newProgram(vertSrc, fragSrc string) (uint32, error) {
	compile := func(src string, shaderType uint32) (uint32, error) {
		s := gl.CreateShader(shaderType)
		csources, free := gl.Strs(src + "\x00")
		defer free()
		gl.ShaderSource(s, 1, csources, nil)
		gl.CompileShader(s)
		var status int32
		gl.GetShaderiv(s, gl.COMPILE_STATUS, &status)
		if status == gl.FALSE {
			var logLen int32
			gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &logLen)
			log := string(make([]byte, logLen))
			gl.GetShaderInfoLog(s, logLen, nil, gl.Str(log+"\x00"))
			return 0, fmt.Errorf("shader compile error: %s", log)
		}
		return s, nil
	}
	vs, err := compile(vertSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fs, err := compile(fragSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)
	var stat int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &stat)
	if stat == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLen)
		log := string(make([]byte, logLen))
		gl.GetProgramInfoLog(prog, logLen, nil, gl.Str(log+"\x00"))
		return 0, fmt.Errorf("program link error: %s", log)
	}
	gl.DeleteShader(vs)
	gl.DeleteShader(fs)
	return prog, nil
}

func loadTextures(paths []string) ([]uint32, error) {
	var ids []uint32
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		rgba := image.NewRGBA(img.Bounds())
		for y := 0; y < rgba.Rect.Dy(); y++ {
			for x := 0; x < rgba.Rect.Dx(); x++ {
				rgba.Set(x, y, img.At(x, y))
			}
		}
		var tex uint32
		gl.GenTextures(1, &tex)
		gl.BindTexture(gl.TEXTURE_2D, tex)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
			int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy()), 0,
			gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
		ids = append(ids, tex)
	}
	return ids, nil
}
