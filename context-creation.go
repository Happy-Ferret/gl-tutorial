package main

import (
	"fmt"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
)

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

func main() {
	var (
		err     error
		monitor *glfw.Monitor
		window  *glfw.Window
		buffers []gl.Buffer
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

	monitor, err = glfw.GetPrimaryMonitor()
	if err != nil {
		panic(err)
	}

	window, err = glfw.CreateWindow(2560, 1600, "Testing", monitor, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	window.SetKeyCallback(handleKey)

	gl.Init()
	buffers = make([]gl.Buffer, 1)
	gl.GenBuffers(buffers)
	fmt.Println(buffers)

	for !window.ShouldClose() {
		//Do OpenGL stuff
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
