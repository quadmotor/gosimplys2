package gosimplys2



import (
	"fmt"
	"time"
	"github.com/udhos/gwob"
)


func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}


func RunObj(input string, output string, target int) {
	defer timer("total run time")()
	options := &gwob.ObjParserOptions{} // parser options

	// o, _ := gwob.NewObjFromFile("bunny.obj", options) // parse/load OBJ
	o, _ := gwob.NewObjFromFile(input, options) //"wall.obj", options) // parse/load OBJ
	fmt.Printf("input total nodes, Triangles %d %d\n", len(o.Coord)/3, len(o.Indices)/3)

	msh := Mesh{}
	nn := o.NumberOfElements()
	nt := len(o.Indices) / 3

	msh.Vertices = make([]Vertex, nn)
	msh.Triangles = make([]Triangle, nt)
	for i := 0; i < nn; i++ {
		x, y, z := o.VertexCoordinates(i)
		msh.Vertices[i].p[0] = float64(x) //float64(o.Coord[i*3+0])
		msh.Vertices[i].p[1] = float64(y) //float64(o.Coord[i*3+1])
		msh.Vertices[i].p[2] = float64(z) //float64(o.Coord[i*3+2])
	}

	for i := 0; i < nt; i++ {
		msh.Triangles[i].v[0] = o.Indices[i*3+0]
		msh.Triangles[i].v[1] = o.Indices[i*3+1]
		msh.Triangles[i].v[2] = o.Indices[i*3+2]
	}

	{
		defer timer("simplify")()
		msh.Simplify(target, 7.0)
	}

	msh.WriteObj(output)
}



