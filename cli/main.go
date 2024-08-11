package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
	"log"
	// "net/http/pprof"
	"runtime/pprof"

	"github.com/quadmotor/gosimplys2"
)

func runTetra() {
	vs := [12]float64{
		0, 0, 0,
		10, 0, 0,
		5, 7, 0,
		5, 3.5, 5,
	}

	tris := [12]int{
		1, 0, 2,
		0, 1, 3,
		1, 2, 3,
		2, 0, 3,
	}

	nn := len(vs) / 3
	nt := len(tris) / 3

	msh := gosimplys2.Mesh{}
	msh.vertices = make([]gosimplys2.Vertex, nn)
	msh.gosimplys2.Triangles = make([]gosimplys2.Triangle, nt)
	for i := 0; i < nn; i++ {
		msh.vertices[i].p[0] = vs[i*3+0]
		msh.vertices[i].p[1] = vs[i*3+1]
		msh.vertices[i].p[2] = vs[i*3+2]
	}

	for i := 0; i < nt; i++ {
		msh.triangles[i].v[0] = tris[i*3+0]
		msh.triangles[i].v[1] = tris[i*3+1]
		msh.triangles[i].v[2] = tris[i*3+2]
	}

	msh.writeObj("input.obj")
	msh.simplify_gosimplys.Mesh(3, 7.0)
	msh.writeObj("output.obj")
}




func main() {

	var input string
	var output string
	var target int

	flag.StringVar(&input, "input", "wall.obj", "input file path")
	flag.StringVar(&output, "output", "result.obj", "output file path")
	flag.IntVar(&target, "target", 2000, "target triangle count")

	//https://go.dev/blog/pprof
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()

	fmt.Printf("input, output, target count %s %s %d\n", input, output, target)

	if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

	if _, err := os.Stat(input); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("input %s does not exist", input)
		return
	}

	gosimplys2.runObj(input, output, target)
}
