package gosimplys2

import (
	"fmt"
	"math"
	"os"
)

type Vertex struct {
	p      vec3f
	tstart int
	tcount int
	q      SymetricMatrix
	border bool
}

type Triangle struct {
	v       [3]int
	err     [4]float64
	n       vec3f
	deleted bool
	dirty   bool
}

type Ref struct {
	tid     int
	tvertex int
}

type Mesh struct {
	Vertices  []Vertex
	Triangles []Triangle
	refs      []Ref
}

func (msh *Mesh) calculate_error(id_v1 int, id_v2 int) float64 {
	q := symAdd(&msh.Vertices[id_v1].q, &msh.Vertices[id_v2].q)
	// border := msh.Vertices[id_v1].border && msh.Vertices[id_v2].border;
	x := msh.Vertices[id_v2].p[0]
	y := msh.Vertices[id_v2].p[1]
	z := msh.Vertices[id_v2].p[2]
	return vertex_error(&q, x, y, z)
}

// Check if a triangle flips when this edge is removed

func (msh *Mesh) flipped(p vec3f, i0 int, i1 int, v0 *Vertex, v1 *Vertex, deleted *[]int) bool {
	bordercount := 0
	for k := 0; k < v0.tcount; k++ {
		t := &msh.Triangles[msh.refs[v0.tstart+k].tid]
		if t.deleted {
			continue
		}

		s := msh.refs[v0.tstart+k].tvertex
		id1 := t.v[(s+1)%3]
		id2 := t.v[(s+2)%3]

		if id1 == i1 || id2 == i1 { // delete ? {
			bordercount++
			(*deleted)[k] = 1
			continue
		}
		d1 := vec3fminus(&msh.Vertices[id1].p, &p)
		vec3fnormalize(&d1)
		d2 := vec3fminus(&msh.Vertices[id2].p, &p)
		vec3fnormalize(&d2)
		if math.Abs(vec3fdot(&d1, &d2)) > 0.999 {
			return true
		}
		n := vec3f{}
		n = vec3fcross(&d1, &d2)
		vec3fnormalize(&n)
		(*deleted)[k] = 0
		if vec3fdot(&n, &t.n) < 0.2 {
			return true
		}
	}
	return false
}

// Update triangle connections and edge error after a edge is collapsed

func (msh *Mesh) update_triangles(i0 int, v *Vertex, deleted *[]int, deleted_triangles *int) {

	for k := 0; k < v.tcount; k++ {
		r := &msh.refs[v.tstart+k]
		t := &msh.Triangles[r.tid]
		if t.deleted {
			continue
		}
		if 1 == (*deleted)[k] {
			t.deleted = true
			*deleted_triangles++
			continue
		}
		t.v[r.tvertex] = i0
		t.dirty = true
		t.err[0] = msh.calculate_error(t.v[0], t.v[1])
		t.err[1] = msh.calculate_error(t.v[1], t.v[2])
		t.err[2] = msh.calculate_error(t.v[2], t.v[0])
		t.err[3] = math.Min(t.err[0], math.Min(t.err[1], t.err[2]))
		msh.refs = append(msh.refs, *r) // msh.refs.push_back(r);
	}
}

func resizeslice(refs *[]int, n3 int) {

	if cap(*refs) > n3 {
		*refs = (*refs)[:n3]
	} else {
		*refs = make([]int, n3)
	}
}

func (msh *Mesh) WriteObj(output string) {
	f, _ := os.Create(output)
	defer f.Close()

	for i := 0; i < len(msh.Vertices); i++ {
		fmt.Fprintf(f, "v %g %g %g\n",
			msh.Vertices[i].p[0], msh.Vertices[i].p[1], msh.Vertices[i].p[2])
	}

	for i := 0; i < len(msh.Triangles); i++ {
		if !msh.Triangles[i].deleted {
			fmt.Fprintf(f, "f %d// %d// %d//\n",
				msh.Triangles[i].v[0]+1, msh.Triangles[i].v[1]+1, msh.Triangles[i].v[2]+1)
		}
	}
}

// compact Triangles, compute edge error and build reference list

func (msh *Mesh) update_mesh(iteration int) {
	if iteration > 0 { // compact Triangles
		dst := 0
		for i := 0; i < len(msh.Triangles); i++ {
			if !msh.Triangles[i].deleted {
				msh.Triangles[dst] = msh.Triangles[i]
				dst++
			}
		}
		msh.Triangles = msh.Triangles[:dst]
	}
	//

	// Init Reference ID list
	// loopi(0,Vertices.size())
	for i := 0; i < len(msh.Vertices); i++ {
		msh.Vertices[i].tstart = 0
		msh.Vertices[i].tcount = 0
	}
	// loopi(0,Triangles.size()){
	for i := 0; i < len(msh.Triangles); i++ {
		t := &msh.Triangles[i]
		for j := 0; j < 3; j++ {
			msh.Vertices[t.v[j]].tcount++
		}
	}
	tstart := 0
	for i := 0; i < len(msh.Vertices); i++ {
		v := &msh.Vertices[i]
		v.tstart = tstart
		tstart += v.tcount
		v.tcount = 0
	}

	// Write References
	n3 := len(msh.Triangles) * 3
	if cap(msh.refs) > n3 {
		msh.refs = msh.refs[:n3]
	} else {
		msh.refs = make([]Ref, n3)
	}

	for i := 0; i < len(msh.Triangles); i++ {
		t := &msh.Triangles[i]
		for j := 0; j < 3; j++ {
			v := &msh.Vertices[t.v[j]]
			msh.refs[v.tstart+v.tcount].tid = i
			msh.refs[v.tstart+v.tcount].tvertex = j
			v.tcount++
		}
	}

	// Init Quadrics by Plane & Edge Errors
	//
	// required at the beginning ( iteration == 0 )
	// recomputing during the simplification is not required,
	// but mostly improves the result for closed meshes
	//
	if iteration == 0 {
		// Identify boundary : Vertices[].border=0,1

		vcount := []int{}
		vids := []int{}

		for i := 0; i < len(msh.Vertices); i++ {
			msh.Vertices[i].border = false
		}

		for i := 0; i < len(msh.Vertices); i++ {
			v := &msh.Vertices[i]
			vcount = vcount[:0]
			vids = vids[:0]
			for j := 0; j < v.tcount; j++ {
				k := msh.refs[v.tstart+j].tid
				t := &msh.Triangles[k]
				for k := 0; k < 3; k++ {
					ofs := 0
					id := t.v[k]
					//while(ofs<vcount.size())
					for {
						if ofs >= len(vcount) {
							break
						}
						if vids[ofs] == id {
							break
						}
						ofs++
					}
					if ofs == len(vcount) {
						vcount = append(vcount, 1)
						vids = append(vids, id)
					} else {
						vcount[ofs]++
					}
				}
			}
			for j := 0; j < len(vcount); j++ {
				if vcount[j] == 1 {
					msh.Vertices[vids[j]].border = true
				}
			}
		}
		//initialize errors
		for i := 0; i < len(msh.Vertices); i++ {
			msh.Vertices[i].q = SymetricMatrix{} //(0.0)
		}

		for i := 0; i < len(msh.Triangles); i++ {

			t := &msh.Triangles[i]
			n := vec3f{}
			p := [3]vec3f{}
			for j := 0; j < 3; j++ {
				p[j] = msh.Vertices[t.v[j]].p
			}

			d01 := vec3fminus(&p[1], &p[0])
			d02 := vec3fminus(&p[2], &p[0])
			n = vec3fcross(&d01, &d02)
			vec3fnormalize(&n)
			t.n = n

			dis := -vec3fdot(&n, &p[0])
			q0 := planeToQuadric(n[0], n[1], n[2], dis)
			for j := 0; j < 3; j++ {
				msh.Vertices[t.v[j]].q = symAdd(&msh.Vertices[t.v[j]].q, &q0)
			}
		}

		for i := 0; i < len(msh.Triangles); i++ {
			// Calc Edge Error
			t := &msh.Triangles[i]
			// p := vec3f{}
			for j := 0; j < 3; j++ {
				t.err[j] = msh.calculate_error(t.v[j], t.v[(j+1)%3])
			}
			t.err[3] = math.Min(t.err[0], math.Min(t.err[1], t.err[2]))
		}
	}
}

// Finally compact mesh before exiting

func (msh *Mesh) compact_mesh() {

	dst := 0

	for i := 0; i < len(msh.Vertices); i++ {
		msh.Vertices[i].tcount = 0
	}

	for i := 0; i < len(msh.Triangles); i++ {
		if !msh.Triangles[i].deleted {
			t := &msh.Triangles[i]
			msh.Triangles[dst] = *t
			dst++
			for j := 0; j < 3; j++ {
				msh.Vertices[t.v[j]].tcount = 1
			}
		}
	}
	msh.Triangles = msh.Triangles[:dst]
	dst = 0

	for i := 0; i < len(msh.Vertices); i++ {
		if msh.Vertices[i].tcount > 0 {
			msh.Vertices[i].tstart = dst
			msh.Vertices[dst].p = msh.Vertices[i].p
			dst++
		}
	}

	for i := 0; i < len(msh.Triangles); i++ {
		t := &msh.Triangles[i]
		for j := 0; j < 3; j++ {
			t.v[j] = msh.Vertices[t.v[j]].tstart
		}
	}
	msh.Vertices = msh.Vertices[:dst]
}

func (msh *Mesh) SetVertex(i int, x float64, y float64, z float64) {
	msh.Vertices[i].p[0] = x
	msh.Vertices[i].p[1] = y
	msh.Vertices[i].p[2] = z
}

func (msh *Mesh) SetTriangle(i int, a int, b int, c int) {
	msh.Triangles[i].v[0] = a
	msh.Triangles[i].v[1] = b
	msh.Triangles[i].v[2] = c
}

// Main simplification function
//
// target_count  : target nr. of Triangles
// agressiveness : sharpness to increase the threshold.
//
//	5..8 are good numbers
//	more iterations yield higher quality
//
// =7
func (msh *Mesh) Simplify(target_count int, agressiveness float64) {
	// init
	// printf("%s - start\n",__FUNCTION__);
	// int timeStart=timeGetTime();

	for i := 0; i < len(msh.Triangles); i++ {
		msh.Triangles[i].deleted = false
	}

	// main iteration loop

	deleted_triangles := 0
	deleted0 := []int{}
	deleted1 := []int{}
	triangle_count := len(msh.Triangles)

	for iteration := 0; iteration < 100; iteration++ {

		// target number of Triangles reached ? Then break
		// printf("iteration %d - Triangles %d\n",iteration,triangle_count-deleted_triangles);
		if triangle_count-deleted_triangles <= target_count {
			break
		}

		// update mesh once in a while
		if iteration%5 == 0 {
			msh.update_mesh(iteration)
		}

		// clear dirty flag
		for i := 0; i < len(msh.Triangles); i++ {
			msh.Triangles[i].dirty = false
		}
		//
		// All Triangles with edges below the threshold will be removed
		//
		// The following numbers works well for most models.
		// If it does not, try to adjust the 3 parameters
		//
		threshold := 0.000000001 * math.Pow(float64(iteration+3), agressiveness)

		// remove Vertices & mark deleted Triangles
		for i := 0; i < len(msh.Triangles); i++ {
			t := &msh.Triangles[i]
			if t.err[3] > threshold {
				continue
			}
			if t.deleted {
				continue
			}
			if t.dirty {
				continue
			}

			for j := 0; j < 3; j++ {
				if t.err[j] < threshold {
					i0 := t.v[j]
					v0 := &msh.Vertices[i0]
					i1 := t.v[(j+1)%3]
					v1 := &msh.Vertices[i1]

					// Border check
					if v0.border != v1.border {
						continue
					}

					// Compute vertex to collapse to
					p := v1.p
					msh.calculate_error(i0, i1)

					resizeslice(&deleted0, v0.tcount)
					resizeslice(&deleted1, v1.tcount)

					// don't remove if flipped
					if msh.flipped(p, i0, i1, v0, v1, &deleted0) {
						continue
					}
					if msh.flipped(p, i1, i0, v1, v0, &deleted1) {
						continue
					}

					// not flipped, so remove edge
					v0.p = p
					v0.q = symAdd(&v1.q, &v0.q)
					tstart := len(msh.refs)

					msh.update_triangles(i0, v0, &deleted0, &deleted_triangles)
					msh.update_triangles(i0, v1, &deleted1, &deleted_triangles)

					tcount := len(msh.refs) - tstart

					if tcount <= v0.tcount {
						// save ram
						if tcount > 0 {
							for k := 0; k < tcount; k++ {
								msh.refs[v0.tstart+k] = msh.refs[tstart+k]
							}
						}
					} else {
						// append
						v0.tstart = tstart
					}

					v0.tcount = tcount
					break
				}
			}
			// done?
			if triangle_count-deleted_triangles <= target_count {
				break
			}
		}

	}

	// clean up mesh
	msh.compact_mesh()

}
