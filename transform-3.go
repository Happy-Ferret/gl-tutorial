package main

import (
	"errors"
	"fmt"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
	glm "github.com/go-gl/mathgl/mgl32"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"time"
)

const vertexSource = `
#version 150

in vec2 position;
in vec3 color;
in vec2 texcoord;

out vec3 Color;
out vec2 Texcoord;

uniform mat4 model;
uniform mat4 view;
uniform mat4 proj;

void main()
{
	Texcoord = texcoord;
	Color = color;
	gl_Position = proj * view * model * vec4(position, 0.0, 1.0);
}
`

const fragmentSource = `
#version 150

in vec3 Color;
in vec2 Texcoord;

out vec4 outColor;

uniform sampler2D texKitten;
uniform sampler2D texPuppy;

void main()
{
	vec4 colKitten = texture(texKitten, Texcoord);
	vec4 colPuppy = texture(texPuppy, Texcoord);
	outColor = mix(colKitten, colPuppy, 0.5);
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

// from github.com/go-gl/example/glfw3/gophercube
func createTexture(r io.Reader) (gl.Texture, error) {
	img, err := png.Decode(r)
	if err != nil {
		return gl.Texture(0), err
	}

	rgbaImg, ok := img.(*image.NRGBA)
	if !ok {
		return gl.Texture(0), errors.New("texture must be an NRGBA image")
	}

	textureId := gl.GenTexture()
	textureId.Bind(gl.TEXTURE_2D)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	// flip image: first pixel is lower left corner
	imgWidth, imgHeight := img.Bounds().Dx(), img.Bounds().Dy()
	data := make([]byte, imgWidth*imgHeight*4)
	lineLen := imgWidth * 4
	dest := len(data) - lineLen
	for src := 0; src < len(rgbaImg.Pix); src += rgbaImg.Stride {
		copy(data[dest:dest+lineLen], rgbaImg.Pix[src:src+rgbaImg.Stride])
		dest -= lineLen
	}
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, imgWidth, imgHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, data)

	return textureId, nil
}

func main() {
	var (
		err               error
		window            *glfw.Window
		vbo, ebo          gl.Buffer
		textures          []gl.Texture
		vertices          []gl.GLfloat
		elements          []gl.GLuint
		vertexShader      gl.Shader
		fragmentShader    gl.Shader
		program           gl.Program
		posAttrib         gl.AttribLocation
		colAttrib         gl.AttribLocation
		texAttrib         gl.AttribLocation
		texKittenLocation gl.UniformLocation
		texPuppyLocation  gl.UniformLocation
		modelLocation     gl.UniformLocation
		viewLocation      gl.UniformLocation
		projLocation      gl.UniformLocation
		vao               gl.VertexArray
		model             glm.Mat4
		view              glm.Mat4
		proj              glm.Mat4
		startTime         time.Time
		diffTime          time.Duration
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
	defer vao.Delete()
	vao.Bind()
	checkError("vertex array object")

	// setup vertex data
	vertices = []gl.GLfloat{
		-0.5, 0.5, 1.0, 0.0, 0.0, 0.0, 1.0, // top left
		0.5, 0.5, 0.0, 1.0, 0.0, 1.0, 1.0, // top right
		0.5, -0.5, 0.0, 0.0, 1.0, 1.0, 0.0, // bottom right
		-0.5, -0.5, 1.0, 1.0, 1.0, 0.0, 0.0, // bottom left
	}
	vbo = gl.GenBuffer()
	defer vbo.Delete()
	vbo.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, int(glh.Sizeof(gl.FLOAT))*len(vertices), vertices, gl.STATIC_DRAW)
	checkError("vertex data")

	// setup element data
	elements = []gl.GLuint{
		0, 1, 2,
		2, 3, 0,
	}
	ebo = gl.GenBuffer()
	defer ebo.Delete()
	ebo.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(glh.Sizeof(gl.UNSIGNED_INT))*len(elements), elements, gl.STATIC_DRAW)
	checkError("element data")

	// setup texture data
	textures = make([]gl.Texture, 2)
	sample, err := os.Open("sample.png")
	if err != nil {
		panic(err)
	}
	gl.ActiveTexture(gl.TEXTURE0)
	textures[0], err = createTexture(sample)
	if err != nil {
		panic(err)
	}
	defer textures[0].Delete()
	sample.Close()

	sample2, err := os.Open("sample2.png")
	if err != nil {
		panic(err)
	}
	gl.ActiveTexture(gl.TEXTURE1)
	textures[1], err = createTexture(sample2)
	if err != nil {
		panic(err)
	}
	defer textures[1].Delete()
	sample2.Close()

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
	posAttrib.AttribPointer(2, gl.FLOAT, false, 7*int(glh.Sizeof(gl.FLOAT)), nil)
	checkError("position attrib pointer")

	// color attribute
	colAttrib = program.GetAttribLocation("color")
	colAttrib.EnableArray()
	colAttrib.AttribPointer(3, gl.FLOAT, false, 7*int(glh.Sizeof(gl.FLOAT)), uintptr(2*int(glh.Sizeof(gl.FLOAT))))
	checkError("color attrib pointer")

	// texcoord attribute
	texAttrib = program.GetAttribLocation("texcoord")
	texAttrib.EnableArray()
	texAttrib.AttribPointer(2, gl.FLOAT, false, 7*int(glh.Sizeof(gl.FLOAT)), uintptr(5*int(glh.Sizeof(gl.FLOAT))))
	checkError("texcoord attrib pointer")

	// setup texture uniforms
	texKittenLocation = program.GetUniformLocation("texKitten")
	texKittenLocation.Uniform1i(0)
	texPuppyLocation = program.GetUniformLocation("texPuppy")
	texPuppyLocation.Uniform1i(1)

	// setup matrices
	modelLocation = program.GetUniformLocation("model")

	viewLocation = program.GetUniformLocation("view")
	view = glm.LookAtV(
		glm.Vec3{1.2, 1.2, 1.2},
		glm.Vec3{0.0, 0.0, 0.0},
		glm.Vec3{0.0, 0.0, 1.0})
	viewLocation.UniformMatrix4fv(false, view)

	projLocation = program.GetUniformLocation("proj")
	proj = glm.Perspective(45.0, 800.0/600.0, 1.0, 10.0)
	projLocation.UniformMatrix4fv(false, proj)

	startTime = time.Now()
	for !window.ShouldClose() {
		glfw.PollEvents()

		// rotate
		diffTime = time.Since(startTime)
		model = glm.HomogRotate3DZ(math.Pi * float32(diffTime.Seconds()))
		modelLocation.UniformMatrix4fv(false, model)

		// clear the screen to black
		width, height := window.GetFramebufferSize()
		gl.Viewport(0, 0, width, height)
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// draw triangles
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)

		checkError("main loop")
		window.SwapBuffers()
	}
}
