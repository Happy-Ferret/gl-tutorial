package main

import (
	"fmt"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
)

const vertexSource = `
#version 150

in vec2 position;
in vec3 color;

out vec3 Color;

void main()
{
	Color = color;
	gl_Position = vec4(position, 0.0, 1.0);
}
`

const fragmentSource = `
#version 150

in vec3 Color;
out vec4 outColor;

void main()
{
	outColor = vec4(Color, 1.0);
}
`

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func handleKey(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press {
		return
	}

	if k == glfw.KeyEscape {
		window.SetShouldClose(true)
	}
}

func checkError(prefix string) {
	if glError := gl.GetError(); glError != gl.NO_ERROR {
		errorString, err := glu.ErrorString(glError)
		if err != nil {
			fmt.Printf("%s: unspecified error!\n", prefix)
		} else {
			fmt.Printf("%s error: %s\n", prefix, errorString)
		}
	}
}

func main() {
	var (
		err            error
		window         *glfw.Window
		vbo            gl.Buffer
		vertices       []gl.GLfloat
		vertexShader   gl.Shader
		fragmentShader gl.Shader
		program        gl.Program
		posAttrib      gl.AttribLocation
		colAttrib      gl.AttribLocation
		vao            gl.VertexArray
	)

	glfw.SetErrorCallback(errorCallback)

	if !glfw.Init() {
		panic("Can't init glfw!")
	}
	defer glfw.Terminate()

	// set opengl version
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
	glfw.WindowHint(glfw.OpenglForwardCompatible, glfw.True)

	// turn off resizing
	glfw.WindowHint(glfw.Resizable, glfw.False)

	window, err = glfw.CreateWindow(800, 600, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()
	window.SetKeyCallback(handleKey)

	gl.Init()
	gl.GetError() // ignore INVALID_ENUM that GLEW raises when using OpenGL 3.2+

	// create Vertex Array Object to save shader attributes
	vao = gl.GenVertexArray()
	vao.Bind()
	checkError("vertex array object")

	// setup vertex data
	vertices = []gl.GLfloat{
		-0.5, 0.5, 1.0, 0.0, 0.0, // top left
		0.5, 0.5, 0.0, 1.0, 0.0, // top right
		0.5, -0.5, 0.0, 0.0, 1.0, // bottom right

		0.5, -0.5, 0.0, 0.0, 1.0, // bottom right
		-0.5, -0.5, 1.0, 1.0, 1.0, // bottom left
		-0.5, 0.5, 1.0, 0.0, 0.0, // top left
	}
	vbo = gl.GenBuffer()
	vbo.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, int(glh.Sizeof(gl.FLOAT))*len(vertices), vertices, gl.STATIC_DRAW)
	checkError("vertex data")

	// compile vertex shader
	vertexShader = gl.CreateShader(gl.VERTEX_SHADER)
	vertexShader.Source(vertexSource)
	vertexShader.Compile()
	if vertexShader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		panic(fmt.Errorf("vertex shader compilation error: %s", vertexShader.GetInfoLog()))
	}
	checkError("vertex shader")

	// compile fragment shader
	fragmentShader = gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentShader.Source(fragmentSource)
	fragmentShader.Compile()
	if fragmentShader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		panic(fmt.Errorf("fragment shader compilation error: %s", fragmentShader.GetInfoLog()))
	}
	checkError("fragment shader")

	// create shader program
	program = gl.CreateProgram()
	program.AttachShader(vertexShader)
	program.AttachShader(fragmentShader)
	program.BindFragDataLocation(0, "outColor")
	program.Link()
	program.Use()
	program.Validate()
	if program.Get(gl.VALIDATE_STATUS) != gl.TRUE {
		panic(fmt.Errorf("program error: %s", program.GetInfoLog()))
	}
	checkError("program")

	// tell vertex shader how to process vertex data
	posAttrib = program.GetAttribLocation("position")
	posAttrib.EnableArray()
	posAttrib.AttribPointer(2, gl.FLOAT, false, 5*int(glh.Sizeof(gl.FLOAT)), nil)
	checkError("position attrib pointer")

	colAttrib = program.GetAttribLocation("color")
	colAttrib.EnableArray()
	colAttrib.AttribPointer(3, gl.FLOAT, false, 5*int(glh.Sizeof(gl.FLOAT)), uintptr(2*int(glh.Sizeof(gl.FLOAT))))
	checkError("color attrib pointer")

	for !window.ShouldClose() {
		glfw.PollEvents()

		// clear the screen to black
		width, height := window.GetFramebufferSize()
		gl.Viewport(0, 0, width, height)
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// draw triangles
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		checkError("main loop")
		window.SwapBuffers()
	}
}
